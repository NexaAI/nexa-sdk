package downloader

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/valyala/fasthttp"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

type HTTPDownloader struct {
	client fasthttp.Client

	token           string
	totalDownloaded atomic.Int64
	progress        chan<- types.DownloadInfo
}

func NewDownloader(progress chan<- types.DownloadInfo) *HTTPDownloader {
	return &HTTPDownloader{
		client: fasthttp.Client{
			NoDefaultUserAgentHeader:  true,
			MaxIdemponentCallAttempts: 3,
			ReadBufferSize:            64 * 1024,
			WriteBufferSize:           64 * 1024,
		},
		progress: progress,
	}
}

func (d *HTTPDownloader) SetToken(token string) {
	d.token = token
}

func (d *HTTPDownloader) Download(ctx context.Context, url, outputPath string) error {
	slog.Debug("Download", "url", url, "outputPath", outputPath)

	url, err := d.handleRedirect(url, 3)
	if err != nil {
		return fmt.Errorf("failed to handle redirect: %v", err)
	}

	contentLength, err := d.getFilesSize(url)
	if err != nil {
		return fmt.Errorf("failed to get file size: %v", err)
	}

	{ // pre-allocate file
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		err = file.Truncate(contentLength)
		file.Close()
		if err != nil {
			return fmt.Errorf("failed to truncate file: %v", err)
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	chunkSize := calcChunkSize(contentLength)
	errCh := make(chan error, 1)

	for start := int64(0); start < contentLength; start += chunkSize {
		slog.Info("Scheduling chunk", "start", start, "contentLength", contentLength)
		end := start + chunkSize - 1
		if end >= contentLength {
			end = contentLength - 1
		}

		select {
		case sem <- struct{}{}:
			wg.Add(1)

			go func(start, end int64) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := d.downloadChunk(cancelCtx, url, outputPath, start, end); err != nil {
					select { // non-blocking send error
					case errCh <- err:
						slog.Error("Download chunk failed", "start", start, "end", end, "error", err)
					default:
					}
					cancel()
				}

				slog.Error("Download chunk done", "start", start, "end", end, "error", err)
			}(start, end)

		case <-cancelCtx.Done():
			wg.Wait()
			return fmt.Errorf("download canceled: %v", cancelCtx.Err())
		}
	}
	wg.Wait()

	select {
	case err := <-errCh:
		os.Remove(outputPath)
		return err
	default:
		return nil
	}
}

func (d *HTTPDownloader) getFilesSize(url string) (int64, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodHead)
	if d.token != "" {
		req.Header.Set("Authorization", "Bearer "+d.token)
	}
	req.Header.Set("Accept-Encoding", "identity")

	if err := d.client.Do(req, resp); err != nil {
		return -1, err
	}

	return int64(resp.Header.ContentLength()), nil
}

func (d *HTTPDownloader) handleRedirect(url string, maxRedirect int) (string, error) {
	currentURL := url

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	for range maxRedirect {
		req.Reset()
		resp.Reset()

		req.SetRequestURI(currentURL)
		req.Header.SetMethod(fasthttp.MethodHead)
		if d.token != "" {
			req.Header.Set("Authorization", "Bearer "+d.token)
		}

		if err := d.client.Do(req, resp); err != nil {
			return "", fmt.Errorf("request failed: %w", err)
		}

		statusCode := resp.StatusCode()
		if statusCode >= 300 && statusCode < 400 {
			location := resp.Header.Peek("Location")
			if len(location) == 0 {
				return "", fmt.Errorf("redirect status %d with no Location", statusCode)
			}
			currentURL = resolveRelativeURL(currentURL, string(location))
			continue
		}

		if statusCode >= 200 && statusCode < 300 {
			return currentURL, nil
		}

		return "", fmt.Errorf("unexpected status code: %d (%s)", statusCode, currentURL)
	}

	return "", fmt.Errorf("exceeded max redirects (%d)", maxRedirect)
}

func resolveRelativeURL(base, location string) string {
	u, err := url.Parse(location)
	if err != nil {
		return location
	}
	if u.IsAbs() {
		return location
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return location
	}
	return baseURL.ResolveReference(u).String()
}

func calcChunkSize(totalSize int64) int64 {
	const (
		minChunkSize int64 = 1 << 20 // 1MB
		maxChunkSize int64 = 8 << 20 // 8MB
		maxChunks          = 100     // max chunks
	)

	chunkSize := totalSize / maxChunks
	if chunkSize < minChunkSize {
		return minChunkSize
	}
	if chunkSize > maxChunkSize {
		return maxChunkSize
	}
	return chunkSize
}

// TODO: ctx not work for fasthttp
func (d *HTTPDownloader) downloadChunk(_ context.Context, url, outputPath string, start, end int64) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	if d.token != "" {
		req.Header.Set("Authorization", "Bearer "+d.token)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	if err := d.client.Do(req, resp); err != nil {
		return err
	}

	if resp.StatusCode() != fasthttp.StatusPartialContent && resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	file, err := os.OpenFile(outputPath, os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.WriteAt(resp.Body(), start)
	if err != nil {
		return err
	}

	if expected := int(end - start + 1); n != expected {
		return fmt.Errorf("write incomplete: wrote %d bytes, expected %d", n, expected)
	}

	d.totalDownloaded.Add(int64(n))
	d.progress <- types.DownloadInfo{
		Downloaded: d.totalDownloaded.Load(),
	}

	return nil
}
