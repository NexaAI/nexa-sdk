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
	hubs = hubs[1:]

	os.Exit(m.Run())
}

func TestModelInfo(t *testing.T) {
	data, _, err := ModelInfo(context.Background(), "NexaAI/OmniNeural-4B")
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestGetFileContent(t *testing.T) {
	data, err := GetFileContent(context.Background(), "NexaAI/OmniNeural-4B", ".gitattributes")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("GetFileContent:\n%s", data)
}

func TestDownload(t *testing.T) {
	files, _, err := ModelInfo(context.Background(), "NexaAI/OmniNeural-4B")
	if err != nil {
		t.Error(err)
		return
	}

	resCh, errCh := StartDownload(context.Background(), "NexaAI/OmniNeural-4B", "/tmp/OmniNeural-4B", files)
	for p := range resCh {
		t.Logf("Downloaded: %d / %d", p.TotalDownloaded, p.TotalSize)
	}
	for e := range errCh {
		t.Error(e)
	}

	os.RemoveAll("/tmp/OmniNeural-4B/")
}

func BenchmarkDownload(b *testing.B) {
	files, _, err := ModelInfo(context.Background(), "ggml-org/embeddinggemma-300M-qat-q4_0-GGUF")
	if err != nil {
		b.Error(err)
		return
	}

	resCh, errCh := StartDownload(context.Background(), "ggml-org/embeddinggemma-300M-qat-q4_0-GGUF", "/tmp/embeddinggemma-300M-qat-q4_0-GGUF", files)
	for p := range resCh {
		b.Logf("Downloaded: %d / %d", p.TotalDownloaded, p.TotalSize)
	}
	for e := range errCh {
		b.Error(e)
	}
}
