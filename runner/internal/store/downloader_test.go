package store

import (
	"context"
	"testing"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

func TestModelInfo(t *testing.T) {
	s := Get()
	data, err := s.ModelInfo(context.Background(), "NexaAI/OmniNeural-4B")
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestFileSize(t *testing.T) {
	s := Get()
	data, err := s.FileSize(context.Background(), "NexaAI/OmniNeural-4B", "weights-2-8.nexa")
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestDownload(t *testing.T) {
	progress := make(chan types.DownloadInfo)
	d := NewDownloader(1165936148, progress)
	go func() {
		for p := range progress {
			t.Logf("Downloaded %d / %d", p.CurrentDownloaded, p.CurrentSize)
		}
	}()
	err := d.Download(context.Background(), "https://huggingface.co/NexaAI/OmniNeural-4B/resolve/main/weights-2-8.nexa", "/tmp/weights-2-8.nexa")
	if err != nil {
		t.Error(err)
	}
}
