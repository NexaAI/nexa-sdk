package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

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

type hfSibling struct {
	RFileName string `json:"rfilename"`
}

type hfModelInfo struct {
	Siblings []hfSibling `json:"siblings"`
}

func (s *Store) HFModelInfo(ctx context.Context, name string) ([]string, error) {
	client := resty.New()
	client.SetResponseMiddlewares(
		httpCodeToError,
		resty.AutoParseResponseMiddleware,
	)
	defer client.Close()

	info := hfModelInfo{}
	_, err := client.R().
		//EnableDebug().
		SetAuthToken(config.Get().HFToken).
		SetResult(&info).
		Get(fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name))
	if err != nil {
		return nil, err
	}

	res := make([]string, len(info.Siblings))
	for i := range info.Siblings {
		res[i] = info.Siblings[i].RFileName
	}

	return res, nil
}

func (s *Store) HFFileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	client := resty.New()
	client.SetResponseMiddlewares(httpCodeToError)
	defer client.Close()

	url := fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(config.Get().HFToken).
		SetHeader("Accept-Encoding", "").
		Head(url)
	if err != nil {
		return -1, err
	}

	length := resp.Header().Get("Content-Length")
	if length == "" {
		return -1, fmt.Errorf("HEAD response missing Content-Length: %s", fileName)
	}

	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid Content-Length: %w, %s", err, fileName)
	}

	return size, nil
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) Pull(ctx context.Context, mf types.ModelManifest) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	if err := s.LockModel(mf.Name); err != nil {
		errC <- err
		close(errC)
		close(infoC)
		return
	}

	go func() {
		defer s.UnlockModel(mf.Name)

		defer close(errC)
		defer close(infoC)

		// filter download file
		var needs []string
		needs = append(needs, mf.ModelFile)
		if mf.MMProjFile != "" {
			needs = append(needs, mf.MMProjFile)
		}
		needs = append(needs, mf.ExtraFiles...)

		// Create model directory structure
		encName := s.encodeName(mf.Name)
		err := os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		pgetDownloader := NewPgetDownloader(mf.Size)
		for _, file := range needs {
			outputPath := path.Join(s.home, "models", encName, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, mf.Name, file)

			err := pgetDownloader.DownloadWithProgress(ctx, downloadURL, config.Get().HFToken, outputPath, infoC)
			if err != nil {
				errC <- err
				return
			}
		}

		model := types.ModelManifest{
			Name:       mf.Name,
			Size:       mf.Size,
			Quant:      mf.Quant,
			ModelFile:  mf.ModelFile,
			MMProjFile: mf.MMProjFile,
			ExtraFiles: mf.ExtraFiles,
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
