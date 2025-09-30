package model_hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
)

const HF_ENDPOINT = "https://huggingface.co"

type HuggingFace struct {
	downloader *downloader.HTTPDownloader
}

func NewHuggingFace() *HuggingFace {
	return &HuggingFace{downloader: downloader.NewDownloader(config.Get().HFToken)}
}

func (d *HuggingFace) CheckAvailable(ctx context.Context, name string) error {
	return nil
}

func (d *HuggingFace) MaxConcurrency() int {
	if config.Get().HFToken != "" {
		return 8
	} else {
		return 1
	}
}

func (d *HuggingFace) ModelInfo(ctx context.Context, name string) ([]ModelFileInfo, error) {
	info := struct {
		Siblings []struct {
			RFileName string `json:"rfilename"`
		} `json:"siblings"`
	}{}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	code, url, err := downloader.FastHTTPResolveRedirect(&d.downloader.Client, config.Get().HFToken, fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name), 3)
	if err != nil {
		if code == 401 || code == 404 {
			return nil, fmt.Errorf("model %s not found on huggingface, please check model id", name)
		}
		return nil, err
	}
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	if config.Get().HFToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Get().HFToken)
	}

	if err := d.downloader.Client.Do(req, resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() == 401 || resp.StatusCode() == 404 {
		return nil, fmt.Errorf("model %s not found on huggingface, please check model id", name)
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("HTTPError: [%d] %s", resp.StatusCode(), string(resp.Body()))
	}

	if err := sonic.Unmarshal(resp.Body(), &info); err != nil {
		return nil, err
	}

	res := make([]ModelFileInfo, len(info.Siblings))
	var error error
	var resLock sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, d.MaxConcurrency())
	for i := range info.Siblings {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			size, err := d.downloader.GetFileSize(fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, name, info.Siblings[i].RFileName))

			resLock.Lock()
			defer resLock.Unlock()

			if err != nil {
				slog.Error("Get file size error", "model", name, "file", info.Siblings[i].RFileName, "err", err)
				error = err
				return
			}
			res[i] = ModelFileInfo{
				Name: info.Siblings[i].RFileName,
				Size: size,
			}
		}(i)
	}
	wg.Wait()

	if error != nil {
		return nil, error
	}

	return res, nil
}

func (d *HuggingFace) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	return d.downloader.DownloadChunk(ctx, fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName), offset, limit, writer)
}
