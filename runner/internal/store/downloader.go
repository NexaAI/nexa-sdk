package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
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
		if err := s.Remove(mf.Name); err != nil {
			errC <- err
			return
		}

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
		if mf.TokenizerFile.Name != "" {
			if mf.TokenizerFile.Downloaded {
				needs = append(needs, mf.TokenizerFile.Name)
			}
		}
		for _, f := range mf.ExtraFiles {
			if f.Downloaded {
				needs = append(needs, f.Name)
			}
		}

		// Create model directory structure
		err := os.MkdirAll(filepath.Join(s.home, "models", mf.Name), 0o770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		downloader := NewHFDownloader(mf.GetSize(), infoC)
		for _, file := range needs {
			outputPath := filepath.Join(s.home, "models", mf.Name, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, mf.Name, file)

			err = downloader.Download(ctx, downloadURL, outputPath)
			if err != nil {
				errC <- err
				return
			}
		}

		// Detect plugin
		name := strings.ToLower(mf.Name)
		switch {
		case strings.Contains(name, "mlx"):
			mf.PluginId = "mlx"
		case strings.Contains(name, "ort-llama"):
			if strings.Contains(name, "cuda") {
				mf.PluginId = "ort_cuda_llama_cpp"
			} else {
				mf.PluginId = "ort_dml_llama_cpp"
			}
		case strings.Contains(name, "ort"):
			if strings.Contains(name, "cuda") {
				mf.PluginId = "ort_cuda"
			} else {
				mf.PluginId = "ort_dml"
			}
		default:
			mf.PluginId = "llama_cpp"
		}

		model := types.ModelManifest{
			Name:          mf.Name,
			ModelType:     mf.ModelType,
			PluginId:      mf.PluginId,
			ModelFile:     mf.ModelFile,
			MMProjFile:    mf.MMProjFile,
			TokenizerFile: mf.TokenizerFile,
			ExtraFiles:    mf.ExtraFiles,
		}
		manifestPath := filepath.Join(s.home, "models", mf.Name, "nexa.manifest")
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
func (s *Store) PullExtraQuant(ctx context.Context, omf, nmf types.ModelManifest) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	go func() {
		defer close(errC)
		defer close(infoC)

		if err := s.LockModel(nmf.Name); err != nil {
			errC <- err
			return
		}
		defer s.UnlockModel(nmf.Name)

		// filter download file
		var needs []string
		for q, f := range nmf.ModelFile {
			if f.Downloaded && !omf.ModelFile[q].Downloaded {
				needs = append(needs, f.Name)
			}
		}
		if nmf.TokenizerFile.Downloaded && !omf.TokenizerFile.Downloaded {
			needs = append(needs, nmf.TokenizerFile.Name)
		}
		for q, f := range nmf.ExtraFiles {
			if f.Downloaded && !omf.ExtraFiles[q].Downloaded {
				needs = append(needs, f.Name)
			}
		}

		// Create model directory structure
		err := os.MkdirAll(filepath.Join(s.home, "models", nmf.Name), 0o770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		downloader := NewHFDownloader(nmf.GetSize(), infoC)
		for _, file := range needs {
			outputPath := filepath.Join(s.home, "models", nmf.Name, file)
			downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, nmf.Name, file)

			err = downloader.Download(ctx, downloadURL, outputPath)
			if err != nil {
				errC <- err
				return
			}
		}

		model := types.ModelManifest{
			Name:          nmf.Name,
			ModelType:     nmf.ModelType,
			PluginId:      nmf.PluginId,
			ModelFile:     nmf.ModelFile,
			MMProjFile:    nmf.MMProjFile,
			TokenizerFile: nmf.TokenizerFile,
			ExtraFiles:    nmf.ExtraFiles,
		}
		manifestPath := filepath.Join(s.home, "models", nmf.Name, "nexa.manifest")
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
