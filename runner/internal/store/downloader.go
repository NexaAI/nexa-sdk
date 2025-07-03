package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/types"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

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

	if err := s.LockModel(name); err != nil {
		errC <- err
		close(errC)
		close(infoC)
		return
	}

	var totalSize int64

	go func() {
		defer s.UnlockModel(name)

		defer close(errC)
		defer close(infoC)

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
			outputPath := path.Join(s.home, "models", encName, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file)

			pgetDownloader := NewPgetDownloader()
			err := pgetDownloader.DownloadWithProgress(ctx, downloadURL, config.Get().HFToken, outputPath, infoC)
			if err != nil {
				errC <- err
				return
			}

			stat, err := os.Stat(outputPath)
			if err != nil {
				errC <- err
				return
			}
			totalSize += stat.Size()
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
