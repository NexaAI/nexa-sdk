package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	"github.com/bytedance/sonic"
	"resty.dev/v3"
)

func (s *Store) ModelInfo(ctx context.Context, name string) ([]string, error) {
	if CheckChinaMainland() {
		return s.VCModelInfo(ctx, name)
	} else {
		return s.HFModelInfo(ctx, name)
	}
}

func (s *Store) FileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	if CheckChinaMainland() {
		return s.VCFileSize(ctx, modelName, fileName)
	} else {
		return s.HFFileSize(ctx, modelName, fileName)
	}
}

func (s *Store) GetQuantInfo(ctx context.Context, modelName string) (int, error) {
	if CheckChinaMainland() {
		return s.VCGetQuantInfo(ctx, modelName)
	} else {
		return s.HFGetQuantInfo(ctx, modelName)
	}
}

var isChinaMainland bool
var checkOnce sync.Once

func CheckChinaMainland() bool {
	checkOnce.Do(func() {
		client := resty.New()
		client.SetTimeout(2 * time.Second)
		defer client.Close()

		for _, ep := range [][]string{
			{"http://ip-api.com/json", "countryCode"},
			{"https://ipapi.co/json", "country_code"},
			{"https://ipinfo.io/json", "country"},
		} {
			res, err := client.R().
				// EnableDebug().
				Get(ep[0])
			if err != nil {
				continue
			}

			n, err := sonic.GetFromString(res.String(), ep[1])
			if err != nil {
				continue
			}

			code, err := n.String()
			if err != nil {
				continue
			}

			slog.Info("Detected country code", "endpoint", ep[0], "code", code)
			isChinaMainland = code == "CN"
			break
		}
	})
	return isChinaMainland
}

type Downloader interface {
	Download(ctx context.Context, url, outputPath string) error
}

func NewDownloader(totalSize int64, progress chan<- types.DownloadInfo) Downloader {
	if CheckChinaMainland() {
		return NewVCDownloader(totalSize, progress)
	} else {
		return NewHFDownloader(totalSize, progress)
	}
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
		downloader := NewDownloader(mf.GetSize(), infoC)
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
		// TODO: abstract this logic to a function and enum plugin_id
		lowerName := strings.ToLower(mf.Name)
		switch {
		case strings.Contains(lowerName, "mlx"):
			mf.PluginId = "mlx"
		case strings.Contains(lowerName, "ort"), strings.Contains(lowerName, "onnx"):
			mf.PluginId = "ort"
		case strings.Contains(lowerName, "npu"), strings.Contains(lowerName, "omni"):
			mf.PluginId = "qnn"
		default:
			mf.PluginId = "llama_cpp"
		}

		model := types.ModelManifest{
			Name:       mf.Name,
			ModelType:  mf.ModelType,
			PluginId:   mf.PluginId,
			ModelFile:  mf.ModelFile,
			MMProjFile: mf.MMProjFile,
			ExtraFiles: mf.ExtraFiles,
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
		downloader := NewDownloader(nmf.GetSize(), infoC)
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
			Name:       nmf.Name,
			ModelType:  nmf.ModelType,
			PluginId:   nmf.PluginId,
			ModelFile:  nmf.ModelFile,
			MMProjFile: nmf.MMProjFile,
			ExtraFiles: nmf.ExtraFiles,
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
