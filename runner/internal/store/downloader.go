package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/types"
	"github.com/bytedance/sonic"
	"resty.dev/v3"
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
	ModelType     types.ModelType
	ModelFile     string
	TokenizerFile string
	ExtraFiles    []string
}

func (s *Store) HFRepoFiles(ctx context.Context, name string) (*types.ModelManifest, []string, error) {
	fileInfos := make([]HFFileInfo, 0)

	client := resty.New()
	client.SetResponseMiddlewares(
		httpCodeToError,
		resty.AutoParseResponseMiddleware,
	)
	defer client.Close()

	_, err := client.R().
		//EnableDebug().
		SetAuthToken(config.Get().HFToken).
		SetResult(&fileInfos).
		Get(fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name))
	if err != nil {
		return nil, nil, err
	}

	var manifest *types.ModelManifest
	filenames := make([]string, 0, len(fileInfos))
	for _, file := range fileInfos {
		// skip dir
		if strings.Contains(file.Path, "/") {
			continue
		}
		if file.Path == "nexa.manifest" {
			// download and parse mf
			mf := types.ModelManifest{}
			res, err := client.R().
				//EnableDebug().
				SetAuthToken(config.Get().HFToken).
				Get(fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path))
			if err != nil {
				return nil, nil, err
			}
			sonic.UnmarshalString(res.String(), &mf)
			manifest = &mf
		}
		filenames = append(filenames, file.Path)
	}

	return manifest, filenames, nil
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) Pull(ctx context.Context, name string, opt PullOption) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	if err := s.TryLockModel(name); err != nil {
		errC <- err
		close(errC)
		close(infoC)
		return
	}

	var totalSize int64

	client := resty.New()
	client.SetResponseMiddlewares(
		httpCodeToError,
		func(c *resty.Client, r *resty.Response) error {
			dInfo := types.DownloadInfo{}
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
					dInfo.CurrentDownloaded += int64(len(b))
					infoC <- dInfo
					return len(b), nil
				},
			}
			return nil
		},
		resty.SaveToFileResponseMiddleware,
	)

	go func() {
		defer s.UnlockModel(name)

		defer close(errC)
		defer close(infoC)
		defer client.Close()

		if !slices.Contains([]types.ModelType{
			types.ModelTypeLLM, types.ModelTypeVLM, types.ModelTypeEmbedder, types.ModelTypeReranker,
		}, opt.ModelType) {
			errC <- fmt.Errorf("not support model type: %s", opt.ModelType)
			return
		}

		// filter download file
		var needs []string
		needs = append(needs, opt.ModelFile)
		if opt.TokenizerFile != "" {
			needs = append(needs, opt.TokenizerFile)
		}
		needs = append(needs, opt.ExtraFiles...)

		// Create model directory structure
		encName := s.encodeName(name)
		err := os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		for _, file := range needs {
			res, err := client.R().
				SetDoNotParseResponse(true).
				SetSaveResponse(true).
				SetOutputFileName(path.Join(s.home, "models", encName, file)).
				SetAuthToken(config.Get().HFToken).
				Get(fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file))
			if err != nil {
				errC <- err
				return
			}
			totalSize += res.Size()
		}

		model := types.ModelManifest{
			Name:          name,
			Size:          totalSize,
			ModelType:     opt.ModelType,
			ModelFile:     opt.ModelFile,
			TokenizerFile: opt.TokenizerFile,
			ExtraFiles:    opt.ExtraFiles,
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

func httpCodeToError(c *resty.Client, r *resty.Response) error {
	if r.StatusCode() >= 400 {
		return fmt.Errorf("HTTPError: [%d] %s", r.StatusCode(), r.String())
	}
	return nil
}
