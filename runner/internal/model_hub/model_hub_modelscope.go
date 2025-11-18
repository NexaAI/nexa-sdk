package model_hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
)

const MS_ENDPOINT = "https://modelscope.cn"

type ModelScope struct {
	downloader *downloader.HTTPDownloader
}

func NewModelScope() *ModelScope {
	return &ModelScope{downloader: downloader.NewDownloader("")}
}

func (d *ModelScope) ChinaMainlandOnly() bool {
	return true
}

func (d *ModelScope) MaxConcurrency() int {
	return 8
}

func (d *ModelScope) CheckAvailable(ctx context.Context, name string) error {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	defer client.Close()

	res, err := client.R().Get(fmt.Sprintf("%s/api/v1/models/%s/revisions", MS_ENDPOINT, name))
	if err != nil || res.StatusCode() != 200 {
		slog.Warn("modelscope check model available error", "model", name, "status_code", res.StatusCode(), "err", err)
		return fmt.Errorf("model %s not found on modelscope, please check model id, err: %s", name, err)
	}

	return nil
}

func (d *ModelScope) ModelInfo(ctx context.Context, name string) ([]ModelFileInfo, error) {
	return d.modelInfo(ctx, name, "")
}

func (d *ModelScope) modelInfo(ctx context.Context, name string, root string) ([]ModelFileInfo, error) {
	info := struct {
		Data struct {
			Files []struct {
				Path string
				Size int64
				Type string
			}
		}
	}{}

	client := resty.New()
	defer client.Close()
	client.SetTimeout(10 * time.Second)

	resp, err := client.R().
		Get(fmt.Sprintf("%s/api/v1/models/%s/repo/files?Root=%s", MS_ENDPOINT, name, root))
	if err != nil || resp.StatusCode() != http.StatusOK {
		slog.Warn("modelscope get model info error", "model", name, "status_code", resp.StatusCode(), "err", err)
		return nil, fmt.Errorf("failed to get model info from modelscope for model %s", name)
	}

	if err := sonic.UnmarshalString(resp.String(), &info); err != nil {
		return nil, err
	}

	res := make([]ModelFileInfo, 0)
	for _, file := range info.Data.Files {
		// blob, tree
		switch file.Type {
		case "tree":
			subFiles, err := d.modelInfo(ctx, name, file.Path)
			if err != nil {
				return nil, err
			}
			res = append(res, subFiles...)
		case "blob":
			res = append(res, ModelFileInfo{
				Name: file.Path,
				Size: file.Size,
			})
		default:
			slog.Warn("modelscope unknown file type", "model", name, "file", file.Path, "type", file.Type)
			continue
		}
	}
	return res, nil
}

func (d *ModelScope) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	return d.downloader.DownloadChunk(ctx, fmt.Sprintf("%s/models/%s/resolve/main/%s", MS_ENDPOINT, modelName, fileName), offset, limit, writer)
}
