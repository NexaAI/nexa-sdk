package model_hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"

	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
)

const HF_ENDPOINT = "https://huggingface.co"

type HuggingFace struct {
	client     *resty.Client
	downloader *downloader.HTTPDownloader
}

func NewHuggingFace() *HuggingFace {
	c := resty.New()
	c.SetResponseMiddlewares(
		// httpCodeToError
		func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() >= 400 {
				return fmt.Errorf("HTTPError: [%d] %s", r.StatusCode(), r.String())
			}
			return nil
		},
		resty.AutoParseResponseMiddleware,
	)

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
	_, err := d.client.R().
		//EnableDebug().
		SetAuthToken(config.Get().HFToken).
		SetResult(&info).
		Get(fmt.Sprintf("%s/api/models/%s", HF_ENDPOINT, name))
	if err != nil {
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

func (d *HuggingFace) fileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	url := fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName)
	resp, err := d.client.R().
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

func (d *HuggingFace) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	url := fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, fileName)
	var rangeStr string
	if offset > 0 {
		if limit > 0 {
			rangeStr = fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1)
		} else {
			rangeStr = fmt.Sprintf("bytes=%d-", offset)
		}
	}
	resp, err := d.client.R().
		SetContext(ctx).
		SetAuthToken(config.Get().HFToken).
		SetHeader("Accept-Encoding", "identity").
		SetHeader("Range", rangeStr).
		Get(url)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}
