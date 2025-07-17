package store

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/types"
)

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
		// EnableDebug().
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

// TODO merge HFModel info
func (s *Store) GetQuantInfo(ctx context.Context, modelName string) (int, error) {
	client := resty.New()
	client.SetResponseMiddlewares(httpCodeToError)
	defer client.Close()

	url := fmt.Sprintf("%s/%s/resolve/main/config.json", HF_ENDPOINT, modelName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(config.Get().HFToken).
		Get(url)
	if err != nil {
		return 0, err
	}

	var info struct {
		QuantizationConfig struct {
			Bits int `json:"bits"`
		} `json:"quantization_config"`
	}
	err = sonic.Unmarshal(resp.Bytes(), &info)
	return info.QuantizationConfig.Bits, err
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

	go func() {
		defer close(errC)
		defer close(infoC)

		// clean before
		//if err := s.Remove(mf.Name); err != nil {
		//	errC <- err
		//	return
		//}

		if err := s.LockModel(mf.Name); err != nil {
			errC <- err
			return
		}
		defer s.UnlockModel(mf.Name)

		// filter download file
		var needs []string
		for _, f := range mf.ModelFile {
			if f.Downloaded {
				needs = append(needs, f.Name)
			}
		}
		if mf.MMProjFile.Name != "" {
			if mf.MMProjFile.Downloaded {
				needs = append(needs, mf.MMProjFile.Name)
			}
		}
		for _, f := range mf.ExtraFiles {
			if f.Downloaded {
				needs = append(needs, f.Name)
			}
		}

		// Create model directory structure
		encName := s.encodeName(mf.Name)
		err := os.MkdirAll(path.Join(s.home, "models", encName), 0o770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		downloader := NewHFDownloader(mf.GetSize(), infoC)
		for _, file := range needs {
			outputPath := path.Join(s.home, "models", encName, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, mf.Name, file)

			err = downloader.Download(ctx, downloadURL, outputPath)
			if err != nil {
				errC <- err
				return
			}
		}

		model := types.ModelManifest{
			Name:       mf.Name,
			ModelFile:  mf.ModelFile,
			MMProjFile: mf.MMProjFile,
			ExtraFiles: mf.ExtraFiles,
		}
		manifestPath := path.Join(s.home, "models", encName, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		err = os.WriteFile(manifestPath, manifestData, 0o664)
		if err != nil {
			errC <- err
			return
		}
	}()

	return
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) PullExtraQuant(ctx context.Context, mf types.ModelManifest) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	go func() {
		defer close(errC)
		defer close(infoC)

		// clean before
		//if err := s.Remove(mf.Name); err != nil {
		//	errC <- err
		//	return
		//}

		if err := s.LockModel(mf.Name); err != nil {
			errC <- err
			return
		}
		defer s.UnlockModel(mf.Name)

		// filter download file
		var needs []string
		for _, f := range mf.ModelFile {
			if f.Downloaded {
				needs = append(needs, f.Name)
			}
		}
		if mf.MMProjFile.Name != "" {
			if mf.MMProjFile.Downloaded {
				needs = append(needs, mf.MMProjFile.Name)
			}
		}
		for _, f := range mf.ExtraFiles {
			if f.Downloaded {
				needs = append(needs, f.Name)
			}
		}

		// Create model directory structure
		encName := s.encodeName(mf.Name)
		err := os.MkdirAll(path.Join(s.home, "models", encName), 0o770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		downloader := NewHFDownloader(mf.GetSize(), infoC)
		for _, file := range needs {
			outputPath := path.Join(s.home, "models", encName, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, mf.Name, file)

			err = downloader.Download(ctx, downloadURL, outputPath)
			if err != nil {
				errC <- err
				return
			}
		}

		model := types.ModelManifest{
			Name:       mf.Name,
			ModelFile:  mf.ModelFile,
			MMProjFile: mf.MMProjFile,
			ExtraFiles: mf.ExtraFiles,
		}
		manifestPath := path.Join(s.home, "models", encName, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		err = os.WriteFile(manifestPath, manifestData, 0o664)
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
