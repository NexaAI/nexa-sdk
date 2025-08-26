package model_hub

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

// Downloader defines the interface for downloading model files
type ModelHub interface {
	CheckAvailable(ctx context.Context, modelName string) error
	ModelInfo(ctx context.Context, modelName string) ([]string, error)
	FileSize(ctx context.Context, modelName, fileName string) (int64, error)
	GetQuantInfo(ctx context.Context, modelName string) (int, error)
	StartDownload(ctx context.Context, modelName, filePath, outputPath string) (chan types.DownloadInfo, chan error)
}

var hubs = []ModelHub{
	NewVocles(),
	NewHuggingFace(),
}

// Download wrap functions

var errNotAvailable = fmt.Errorf("no available model hub")

func ModelInfo(ctx context.Context, modelName string) ([]string, error) {
	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", hub, "err", err)
			continue
		}
		return hub.ModelInfo(ctx, modelName)
	}
	return nil, errNotAvailable
}

func FileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", hub, "err", err)
			continue
		}
		return hub.FileSize(ctx, modelName, fileName)
	}
	return 0, errNotAvailable
}

func GetQuantInfo(ctx context.Context, modelName string) (int, error) {
	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", hub, "err", err)
			continue
		}
		return hub.GetQuantInfo(ctx, modelName)
	}
	return 0, errNotAvailable
}

func StartDownload(ctx context.Context, modelName, filePath, outputPath string) (resChan chan types.DownloadInfo, errChan chan error) {
	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", hub, "err", err)
			continue
		}
		return hub.StartDownload(ctx, modelName, filePath, outputPath)
	}

	resCh := make(chan types.DownloadInfo)
	close(resCh)
	errCh := make(chan error, 1)
	errCh <- errNotAvailable
	close(errCh)
	return resCh, errCh
}
