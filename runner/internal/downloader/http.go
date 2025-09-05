package downloader

import (
	"context"
	"io"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/valyala/fasthttp"
)

type HTTPDownloader struct {
	fasthttp.Client
}

func NewDownloader() *HTTPDownloader {
	return &HTTPDownloader{
		Client: fasthttp.Client{
			NoDefaultUserAgentHeader:  true,
			MaxIdemponentCallAttempts: 3,
			ReadBufferSize:            64 * 1024,
			WriteBufferSize:           64 * 1024,
		},
	}
}

func (d *HTTPDownloader) Download(ctx context.Context, url string, offset, limit int64, writer io.Writer) error {
	// 	slog.Debug("Download", "url", url, "outputPath", outputPath)

	// 	url, err := d.handleRedirect(url, 3)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to handle redirect: %v", err)
	// 	}

	// 	contentLength, err := d.GetFilesSize(url)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to get file size: %v", err)
	// 	}

	// 	{ // pre-allocate file
	// 		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
	// 			return fmt.Errorf("failed to create directory: %v", err)
	// 		}
	// 		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0o644)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to create file: %v", err)
	// 		}
	// 		err = file.Truncate(contentLength)
	// 		file.Close()
	// 		if err != nil {
	// 			return fmt.Errorf("failed to truncate file: %v", err)
	// 		}
	// 	}

	// 	var wg sync.WaitGroup
	// 	sem := make(chan struct{}, 8)
	// 	cancelCtx, cancel := context.WithCancel(ctx)
	// 	defer cancel()
	// 	chunkSize := calcChunkSize(contentLength)
	// 	errCh := make(chan error, 1)

	// 	for start := int64(0); start < contentLength; start += chunkSize {
	// 		slog.Info("Scheduling chunk", "start", start, "contentLength", contentLength)
	// 		end := start + chunkSize - 1
	// 		if end >= contentLength {
	// 			end = contentLength - 1
	// 		}

	// 		select {
	// 		case sem <- struct{}{}:
	// 			wg.Add(1)

	// 			go func(start, end int64) {
	// 				defer wg.Done()
	// 				defer func() { <-sem }()
	// 				if err := d.DownloadChunk(cancelCtx, url, outputPath, start, end, progress); err != nil {
	// 					select { // non-blocking send error
	// 					case errCh <- err:
	// 						slog.Error("Download chunk failed", "start", start, "end", end, "error", err)
	// 					default:
	// 					}
	// 					cancel()
	// 				}
	// 			}(start, end)

	// 		case <-cancelCtx.Done():
	// 			wg.Wait()
	// 			return fmt.Errorf("download canceled: %v", cancelCtx.Err())
	// 		}
	// 	}
	// 	wg.Wait()

	// select {
	// case err := <-errCh:
	//
	//	os.Remove(outputPath)
	//	return err
	//
	// default:
	//
	//		return nil
	//	}
	panic("unimplemented")
}

func (d *HTTPDownloader) GetFileSize(url string) (int64, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodHead)
	if config.Get().HFToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Get().HFToken)
	}
	req.Header.Set("Accept-Encoding", "identity")

	if err := d.Client.Do(req, resp); err != nil {
		return -1, err
	}

	return int64(resp.Header.ContentLength()), nil
}
