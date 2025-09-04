package model_hub

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))

	// only test huggingface
	// hubs = hubs[1:]

	os.Exit(m.Run())
}

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
	files := []string{"weights-1-8.nexa", "weights-2-8.nexa", "weights-3-8.nexa", "weights-4-8.nexa", "weights-5-8.nexa", "weights-6-8.nexa", "weights-7-8.nexa", "weights-8-8.nexa"}
	resCh, errCh := StartDownload(context.Background(), "NexaAI/OmniNeural-4B", "/tmp/OmniNeural-4B", files)
	for p := range resCh {
		t.Logf("Downloaded: %d / %d", p.TotalDownloaded, p.TotalSize)
	}
	for e := range errCh {
		t.Error(e)
	}
	for _, f := range files {
		os.Remove("/tmp/OmniNeural-4B/" + f)
	}
}
