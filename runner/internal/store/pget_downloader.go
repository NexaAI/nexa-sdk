package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/NexaAI/nexa-sdk/internal/types"
	"github.com/replicate/pget/pkg/client"
	"github.com/replicate/pget/pkg/download"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

type PgetDownloader struct {
	options download.Options
}

func NewPgetDownloader(n int) *PgetDownloader {
	httpClient := &http.Client{}

	clientOpts := client.Options{
		MaxRetries: 2,
		Transport:  httpClient.Transport,
	}

	opts := download.Options{
		MaxConcurrency: n,
		ChunkSize:      16 << 20, // 16 MiB
		Client:         clientOpts,
	}

	return &PgetDownloader{options: opts}
}

func (pd *PgetDownloader) DownloadWithProgress(ctx context.Context, url, outputPath, authToken string, progressCh chan<- types.DownloadInfo) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if authToken != "" {
		pd.options.Client.Transport = &authTransport{
			Base:      http.DefaultTransport,
			AuthToken: authToken,
		}
	}

	strategy := download.GetBufferMode(pd.options)

	reader, fileSize, err := strategy.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch file info: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	var downloaded int64
	progressReader := &progressReader{
		reader:     reader,
		totalSize:  fileSize,
		downloaded: &downloaded,
		filename:   filepath.Base(outputPath),
		progressCh: progressCh,
	}

	_, err = io.Copy(outputFile, progressReader)
	if err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to download file: %w", err)
	}

	if progressCh != nil {
		progressCh <- types.DownloadInfo{
			CurrentSize:       fileSize,
			CurrentDownloaded: fileSize,
			CurrentName:       filepath.Base(outputPath),
		}
	}

	return nil
}

type authTransport struct {
	Base      http.RoundTripper
	AuthToken string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.AuthToken)
	}
	return t.Base.RoundTrip(req)
}

type progressReader struct {
	reader     io.Reader
	totalSize  int64
	downloaded *int64
	filename   string
	progressCh chan<- types.DownloadInfo
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 && pr.progressCh != nil {
		downloaded := atomic.AddInt64(pr.downloaded, int64(n))

		if downloaded%(1024*1024) == 0 || downloaded == pr.totalSize {
			pr.progressCh <- types.DownloadInfo{
				CurrentSize:       pr.totalSize,
				CurrentDownloaded: downloaded,
				CurrentName:       pr.filename,
			}
		}
	}
	return n, err
}
