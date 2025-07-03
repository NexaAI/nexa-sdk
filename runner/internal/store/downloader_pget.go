package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/replicate/pget/pkg/client"
	"github.com/replicate/pget/pkg/download"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

type PgetDownloader struct {
	options download.Options
}

func NewPgetDownloader() *PgetDownloader {
	return &PgetDownloader{options: download.Options{
		MaxConcurrency: 8,
		ChunkSize:      16 << 20, // 16 MiB
		Client: client.Options{
			MaxRetries: 2,
			Transport:  &authTransport{Base: http.DefaultTransport},
		},
	}}
}

func (pd *PgetDownloader) DownloadWithProgress(ctx context.Context, url, outputPath, authToken string, progressCh chan<- types.DownloadInfo) error {
	pd.options.Client.Transport.(*authTransport).AuthToken = authToken

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	strategy := download.GetBufferMode(pd.options)
	reader, fileSize, err := strategy.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch file info: %w", err)
	}

	var downloaded int64

	_, err = io.Copy(outputFile, &types.TeeReadF{
		Raw: reader,
		WriterF: func(p []byte) (int, error) {
			downloaded += int64(len(p))
			// TODO: reduce
			progressCh <- types.DownloadInfo{
				CurrentSize:       fileSize,
				CurrentDownloaded: downloaded,
				CurrentName:       filepath.Base(outputPath),
			}
			return len(p), nil
		},
	})
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
