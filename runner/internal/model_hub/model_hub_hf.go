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
	return &HuggingFace{downloader: downloader.NewDownloader()}
}

func (d *HuggingFace) CheckAvailable(ctx context.Context, name string) error {
	return nil
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

	url, err := downloader.FastHTTPResolveRedirect(&d.downloader.Client, fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name), 3)
	if err != nil {
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

	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("HTTPError: [%d] %s", resp.StatusCode(), string(resp.Body()))
	}

	if err := sonic.Unmarshal(resp.Body(), &info); err != nil {
		return nil, err
	}

	res := make([]ModelFileInfo, len(info.Siblings))
	var resLock sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)
	for i := range info.Siblings {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			size, err := d.downloader.GetFileSize(fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, name, info.Siblings[i].RFileName))
			if err != nil {
				slog.Error("Get file size error", "model", name, "file", info.Siblings[i].RFileName, "err", err)
				return
			}

			resLock.Lock()
			defer resLock.Unlock()
			res[i] = ModelFileInfo{
				Name: info.Siblings[i].RFileName,
				Size: size,
			}
		}(i)
	}
	wg.Wait()

	// check res
	for _, info := range res {
		if info.Name == "" {
			return nil, fmt.Errorf("failed to get file size: %s", info.Name)
		}
	}

	return res, nil
}

func (d *HuggingFace) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	return d.downloader.DownloadChunk(ctx, fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName), offset, limit, writer)
}
