package downloader

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/valyala/fasthttp"
)

type HTTPDownloader struct {
	authToken    string
	maxRedirects int

	fasthttp.Client
}

func NewDownloader(authToken string) *HTTPDownloader {
	return &HTTPDownloader{
		authToken:    authToken,
		maxRedirects: 3,
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

	for range d.maxRedirects {
		req.Reset()
		resp.Reset()

		req.SetRequestURI(url)
		req.Header.SetMethod(fasthttp.MethodGet)
		req.Header.Set("User-Agent", "NexaSDK/0.0")
		if d.authToken != "" {
			req.Header.Set("Authorization", "Bearer "+d.authToken)
		}
		if limit > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1))
		} else {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		}

		if err := d.Client.Do(req, resp); err != nil {
			return err
		}

		if resp.StatusCode() >= 300 && resp.StatusCode() < 400 {
			location := resp.Header.Peek("Location")
			if len(location) == 0 {
				return fmt.Errorf("redirect status %d with no Location", resp.StatusCode())
			}
			url = resolveRelativeURL(url, string(location))
			continue
		}

		if resp.StatusCode() != fasthttp.StatusPartialContent && resp.StatusCode() != fasthttp.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		_, err := writer.Write(resp.Body())
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("exceeded max redirects (%d)", d.maxRedirects)
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
