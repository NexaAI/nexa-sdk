package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
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

type PullOption struct {
	ModelType types.ModelType
	Model     string
	Tokenizer string
	Extra     []string
	ALl       bool // download all file
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) Pull(ctx context.Context, name string, opt PullOption) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
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
			if r.Request.OutputFileName != "" {
				dInfo.CurrentName = path.Base(r.Request.OutputFileName)
			} else {
				dInfo.CurrentName = "filelist"
			}
			r.Body = &types.TeeReadCloserF{
				Raw: r.Body,
				// TODO: reduce channel message
				WriterF: func(b []byte) (int, error) {
					dInfo.TotalDownloaded += int64(len(b))
					dInfo.CurrentDownloaded += int64(len(b))
					infoC <- dInfo
					return len(b), nil
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

		if !slices.Contains([]types.ModelType{
			types.ModelTypeLLM, types.ModelTypeVLM, types.ModelTypeEmbedder, types.ModelTypeReranker,
		}, opt.ModelType) {
			errC <- fmt.Errorf("not support model type: %s", opt.ModelType)
			return
		}
		if opt.Model == "" {
			opt.ALl = true
		}

		fileInfos := make([]HFFileInfo, 0)
		_, err := client.R().
			//EnableDebug().
			SetAuthToken(config.Get().HFToken).
			SetResult(&fileInfos).
			Get(fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name))
		if err != nil {
			errC <- err
			return
		}

		// filter download file
		var needs []string
		if opt.Model != "" {
			needs = append(opt.Extra, opt.Model)
		}
		if opt.Tokenizer != "" {
			needs = append(needs, opt.Tokenizer)
		}

		if !opt.ALl {
			res := make([]HFFileInfo, 0, 2)

			for _, file := range fileInfos {
				if slices.Contains(needs, file.Path) {
					res = append(res, file)
				}
			}
			fileInfos = res
		}

		// check model and tokenizer exist
		exist := 0
		hasManifest := false
		for _, file := range fileInfos {
			if slices.Contains(needs, file.Path) {
				exist += 1
			}
			if !hasManifest && file.Path == "nexa.manifest" {
				hasManifest = true
			}
			totalSize += file.Size
		}
		if exist != len(needs) {
			errC <- fmt.Errorf("some files not found on huggingface repo, check you file name")
			return
		}

		// Create model directory structure
		encName := s.encodeName(name)
		err = os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		for _, file := range fileInfos {
			// skip subdir
			if path.Base(file.Path) != file.Path {
				errC <- fmt.Errorf("not support subdir: %s", file.Path)
				return
			}
			_, err = client.R().
				SetDoNotParseResponse(true).
				SetSaveResponse(true).
				SetOutputFileName(path.Join(s.home, "models", encName, file.Path)).
				SetAuthToken(config.Get().HFToken).
				Get(fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path))
			if err != nil {
				errC <- err
				return
			}
		}

		// Create and save model manifest with metadata
		if !hasManifest {
			// detect main model file when not specify
			if opt.Model == "" {
				for _, f := range fileInfos {
					// Find first npz then first .gguf file in the model
					if opt.Model == "" {
						if strings.HasSuffix(f.Path, ".gguf") || strings.HasSuffix(f.Path, ".npz") {
							opt.Model = f.Path
						}
					}
					if strings.HasSuffix(f.Path, "npz") && strings.HasSuffix(opt.Model, ".gguf") {
						opt.Model = f.Path
					}
				}
			}

			model := types.Model{
				Name:          name,
				Size:          totalSize,
				ModelType:     opt.ModelType,
				ModelFile:     opt.Model,
				TokenizerFile: opt.Tokenizer,
			}
			manifestPath := path.Join(s.home, "models", encName, "nexa.manifest")
			manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
			err = os.WriteFile(manifestPath, manifestData, 0664)
			if err != nil {
				errC <- err
				return
			}
		}
	}()

	return
}
