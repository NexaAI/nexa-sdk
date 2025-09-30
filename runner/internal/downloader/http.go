package downloader

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/valyala/fasthttp"
)

type HTTPDownloader struct {
	authToken string
	fasthttp.Client
}

func NewDownloader(authToken string) *HTTPDownloader {
	return &HTTPDownloader{
		authToken: authToken,
		Client: fasthttp.Client{
			NoDefaultUserAgentHeader:  true,
			MaxIdemponentCallAttempts: 3,
			ReadBufferSize:            64 * 1024,
			WriteBufferSize:           64 * 1024,
		},
	}
}

func (d *HTTPDownloader) DownloadChunk(ctx context.Context, url string, offset, limit int64, writer io.Writer) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	_, currentURL, err := FastHTTPResolveRedirect(&d.Client, d.authToken, url, 3)
	if err != nil {
		return err
	}

	req.SetRequestURI(currentURL)
	req.Header.SetMethod(fasthttp.MethodGet)
	if d.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.authToken)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1))

	if err := d.Client.Do(req, resp); err != nil {
		return err
	}

	if resp.StatusCode() != fasthttp.StatusPartialContent && resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	_, err = writer.Write(resp.Body())
	if err != nil {
		return err
	}

	return nil
}

func (d *HTTPDownloader) GetFileSize(url string) (int64, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	_, currentURL, err := FastHTTPResolveRedirect(&d.Client, d.authToken, url, 3)
	if err != nil {
		return -1, err
	}

	req.SetRequestURI(currentURL)
	req.Header.SetMethod(fasthttp.MethodHead)
	if d.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+d.authToken)
	}
	req.Header.Set("Accept-Encoding", "identity")

	if err := d.Client.Do(req, resp); err != nil {
		return -1, err
	}

	return int64(resp.Header.ContentLength()), nil
}

func FastHTTPResolveRedirect(client *fasthttp.Client, authToken string, currentURL string, maxRedirect int) (int, string, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	for range maxRedirect {
		req.Reset()
		resp.Reset()

		req.SetRequestURI(currentURL)
		req.Header.SetMethod(fasthttp.MethodHead)
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}

		if err := client.Do(req, resp); err != nil {
			return 0, "", fmt.Errorf("request failed: %w", err)
		}

		statusCode := resp.StatusCode()
		if statusCode >= 300 && statusCode < 400 {
			location := resp.Header.Peek("Location")
			if len(location) == 0 {
				return 0, "", fmt.Errorf("redirect status %d with no Location", statusCode)
			}
			currentURL = resolveRelativeURL(currentURL, string(location))
			req.Reset()
			resp.Reset()
			continue
		}

		if statusCode >= 200 && statusCode < 300 {
			return statusCode, currentURL, nil
		}

		return statusCode, "", fmt.Errorf("unexpected status code: %d (%s)", statusCode, currentURL)
	}

	return 0, "", fmt.Errorf("exceeded max redirects (%d)", maxRedirect)
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
