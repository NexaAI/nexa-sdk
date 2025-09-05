package model_hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
)

const HF_ENDPOINT = "https://huggingface.co"

type HuggingFace struct {
	client     *fasthttp.Client
	downloader *downloader.HTTPDownloader
}

func NewHuggingFace() *HuggingFace {
	c := &fasthttp.Client{
		NoDefaultUserAgentHeader:  true,
		MaxIdemponentCallAttempts: 3,
		ReadBufferSize:            64 * 1024,
		WriteBufferSize:           64 * 1024,
	}

	d := downloader.NewDownloader()

	return &HuggingFace{client: c, downloader: d}
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

	req.SetRequestURI(fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name))
	req.Header.SetMethod(fasthttp.MethodGet)
	if config.Get().HFToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Get().HFToken)
	}

	if err := d.client.Do(req, resp); err != nil {
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
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			size, err := d.fileSize(ctx, name, info.Siblings[i].RFileName)
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
		}()
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

func (d *HuggingFace) fileSize(_ context.Context, modelName, fileName string) (int64, error) {
	url := fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodHead)
	if config.Get().HFToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Get().HFToken)
	}
	req.Header.Set("Accept-Encoding", "")

	if err := d.client.Do(req, resp); err != nil {
		return -1, err
	}

	length := string(resp.Header.Peek("Content-Length"))
	if length == "" {
		return -1, fmt.Errorf("HEAD response missing Content-Length: %s", fileName)
	}

	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid Content-Length: %w, %s", err, fileName)
	}

	return size, nil
}

func (d *HuggingFace) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName))
	req.Header.SetMethod(fasthttp.MethodGet)
	if config.Get().HFToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Get().HFToken)
	}

	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1))
	} else if limit > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", limit-1))
	}

	if err := d.client.Do(req, resp); err != nil {
		return err
	}

	_, err := writer.Write(resp.Body())
	return err
}
