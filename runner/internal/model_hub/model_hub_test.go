package model_hub

import (
	"context"
	"os"
	"testing"
)

func TestModelInfo(t *testing.T) {
	data, err := ModelInfo(context.Background(), "NexaAI/OmniNeural-4B")
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestFileSize(t *testing.T) {
	data, err := FileSize(context.Background(), "NexaAI/OmniNeural-4B", "weights-2-8.nexa")
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestDownload(t *testing.T) {
	hubs = hubs[1:] // only test huggingface
	resCh, errCh := StartDownload(context.Background(), "NexaAI/OmniNeural-4B", "weights-2-8.nexa", "/tmp/weights-2-8.nexa")
	for p := range resCh {
		t.Logf("Downloaded %d / %d", p.Downloaded, 1165936148)
	}
	for e := range errCh {
		t.Error(e)
	}
	os.Remove("/tmp/weights-2-8.nexa")
}
