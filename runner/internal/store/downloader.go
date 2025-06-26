package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/types"
)

const HF_ENDPOINT = "https://huggingface.co"

// HFFileInfo represents file metadata from HuggingFace API response
type HFFileInfo struct {
	Type string `json:"type"`
	OId  string `json:"oid"`
	Size uint64 `json:"size"`
	Path string `json:"path"`
	LFS  struct {
		OId string `json:"oid"`
	} `json:"lfs"`
	XetHash string `json:"xetHash"`
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// TODO: multi gguf file support
func (s *Store) Pull(ctx context.Context, name string) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {

	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	go func() {
		client := resty.New()

		defer close(errC)
		defer close(infoC)
		defer client.Close()

		files := make([]HFFileInfo, 0)
		_, err := client.R().
			EnableDebug().
			SetAuthToken(config.Get().HFToken).
			SetResult(&files).
			Get(fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name))
		if err != nil {
			errC <- err
			return
		}

		dInfo := types.DownloadInfo{}
		var mainName string
		for _, f := range files {
			dInfo.Size += f.Size

			// Find first npz then first .gguf file in the model
			f.Path = path.Base(f.Path)
			if mainName == "" {
				if strings.HasSuffix(f.Path, ".gguf") || strings.HasSuffix(f.Path, ".npz") {
					mainName = f.Path
				}
			}
			if strings.HasSuffix(f.Path, "npz") && strings.HasSuffix(mainName, ".gguf") {
				mainName = f.Path
			}
		}

		// Create model directory structure
		encName := s.encodeName(name)
		e := os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if e != nil {
			errC <- e
			return
		}

		client.SetResponseMiddlewares(
			func(c *resty.Client, r *resty.Response) error {
				r.Body = &types.FuncReadCloser{
					Raw: r.Body,
					// TODO: reduce channel message
					F: func(b []byte) {
						dInfo.Downloaded += uint64(len(b))
						infoC <- dInfo
					},
				}
				return nil
			},
			resty.SaveToFileResponseMiddleware,
		)

		// Create modelfile for storing downloaded content
		for _, file := range files {
			_, err = client.R().
				SetSaveResponse(true).
				SetOutputFileName(path.Join(s.home, "models", encName, path.Base(file.Path))).
				SetAuthToken(config.Get().HFToken).
				Get(fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path))
			if err != nil {
				errC <- e
				return
			}
		}

		// Create and save model manifest with metadata
		model := types.Model{
			Name:      name,
			Size:      dInfo.Size,
			ModelFile: mainName,
		}
		manifestPath := path.Join(s.home, "models", encName, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		e = os.WriteFile(manifestPath, manifestData, 0664)
		if e != nil {
			errC <- e
			return
		}
	}()

	return
}
