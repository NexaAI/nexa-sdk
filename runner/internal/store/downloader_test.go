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
	s := Get()
	infoCh, errCh := s.Pull(context.Background(), types.ModelManifest{
		Name: "NexaAI/OmniNeural-4B",
	})
	for {
		select {
		case info, ok := <-infoCh:
			if !ok {
				infoCh = nil
				continue
			}
			t.Log(info)
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			t.Error(err)
		}
		if infoCh == nil && errCh == nil {
			break
		}
	}
}
