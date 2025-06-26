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
	Size int64  `json:"size"`
	Path string `json:"path"`
	LFS  struct {
		OId string `json:"oid"`
	} `json:"lfs"`
	XetHash string `json:"xetHash"`
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
func (s *Store) Pull(ctx context.Context, name string) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {

	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	var totalSize int64
	client := resty.New()
	client.SetResponseMiddlewares(
		func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() >= 400 {
				return fmt.Errorf("HTTPError: [%d] %s", r.StatusCode(), r.String())
			}
			return nil
		},
		func(c *resty.Client, r *resty.Response) error {
			dInfo := types.DownloadInfo{}
			dInfo.TotalSize = totalSize
			dInfo.CurrentSize = r.RawResponse.ContentLength
			dInfo.CurrentName = path.Base(r.Request.OutputFileName)
			r.Body = &types.FuncReadCloser{
				Raw: r.Body,
				// TODO: reduce channel message
				F: func(b []byte) {
					dInfo.TotalDownloaded += int64(len(b))
					dInfo.CurrentDownloaded += int64(len(b))
					infoC <- dInfo
				},
			}
			return nil
		},
		resty.AutoParseResponseMiddleware,
		resty.SaveToFileResponseMiddleware,
	)

	go func() {
		defer close(errC)
		defer close(infoC)
		defer client.Close()

		files := make([]HFFileInfo, 0)
		_, err := client.R().
			//EnableDebug().
			SetAuthToken(config.Get().HFToken).
			SetResult(&files).
			Get(fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name))
		if err != nil {
			errC <- err
			return
		}

		var mainName string
		for _, f := range files {
			totalSize += f.Size

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
		err = os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		for _, file := range files {
			_, err = client.R().
				SetDoNotParseResponse(true).
				SetSaveResponse(true).
				SetOutputFileName(path.Join(s.home, "models", encName, path.Base(file.Path))).
				SetAuthToken(config.Get().HFToken).
				Get(fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path))
			if err != nil {
				errC <- err
				return
			}
		}

		// Create and save model manifest with metadata
		model := types.Model{
			Name:      name,
			Size:      totalSize,
			ModelFile: mainName,
		}
		manifestPath := path.Join(s.home, "models", encName, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		err = os.WriteFile(manifestPath, manifestData, 0664)
		if err != nil {
			errC <- err
			return
		}
	}()

	return
}
