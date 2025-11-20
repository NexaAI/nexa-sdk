package model_hub

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

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

func (d *HuggingFace) ChinaMainlandOnly() bool {
	return false
}

func (d *HuggingFace) MaxConcurrency() int {
	if config.Get().HFToken != "" {
		return 8
	} else {
		return 1
	}
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

	client := resty.New()
	defer client.Close()
	client.SetTimeout(10 * time.Second)
	client.AddResponseMiddleware(code2error)

	req := client.R()
	if config.Get().HFToken != "" {
		req.SetHeader("Authorization", "Bearer "+config.Get().HFToken)
	}
	resp, err := req.Get(fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name))
	if err != nil {
		return nil, err
	}

	if err := sonic.UnmarshalString(resp.String(), &info); err != nil {
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

			req := client.R()
			if config.Get().HFToken != "" {
				req.SetHeader("Authorization", "Bearer "+config.Get().HFToken)
			}
			req.SetHeader("Accept-Encoding", "identity")
			resp, err := req.Head(fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, name, info.Siblings[i].RFileName))
			resLock.Lock()
			defer resLock.Unlock()

			if err != nil {
				error = err
				return
			}
			if resp.StatusCode() != http.StatusOK || resp.RawResponse.ContentLength < 0 {
				error = fmt.Errorf("Get file [%s] info error: %s", info.Siblings[i].RFileName, resp.Status())
				return
			}
			res[i] = ModelFileInfo{
				Name: info.Siblings[i].RFileName,
				Size: resp.RawResponse.ContentLength,
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
