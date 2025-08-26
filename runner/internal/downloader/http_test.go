package downloader

import (
	"context"
	"testing"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

func TestDownloader(t *testing.T) {
	bar := render.NewProgressBar(1165936148, "")
	progress := make(chan types.DownloadInfo, 10)
	d := NewDownloader(progress)
	d.Download(context.Background(), "https://huggingface.co/NexaAI/OmniNeural-4B/resolve/main/weights-2-8.nexa", "/tmp/weights-2-8.nexa")
	for pg := range progress {
		bar.Set(pg.Downloaded)
	}
	bar.Exit()
}
