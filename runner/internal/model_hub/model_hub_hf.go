package model_hub

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
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

func (d *HuggingFace) ModelInfo(ctx context.Context, name string) ([]string, error) {
	info := struct {
		Siblings []struct {
			RFileName string `json:"rfilename"`
		} `json:"siblings"`
	}{}
	_, err := d.client.R().
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

func (d *HuggingFace) FileSize(ctx context.Context, modelName, fileName string) (int64, error) {
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

// TODO merge HFModel info
func (d *HuggingFace) GetQuantInfo(ctx context.Context, modelName string) (int, error) {
	url := fmt.Sprintf("%s/%s/resolve/main/config.json", HF_ENDPOINT, modelName)
	resp, err := d.client.R().
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

func (d *HuggingFace) StartDownload(ctx context.Context, modelName, outputPath string, files []string) (chan types.DownloadInfo, chan error) {
	d.downloader.SetToken(config.Get().HFToken)

	resCh := make(chan types.DownloadInfo, 8)
	errCh := make(chan error, 1)
	go func() {
		progressCh := make(chan int64, 8)
		defer close(progressCh)

		go func() {
			defer close(errCh)
			defer close(resCh)

			var info types.DownloadInfo
			for p := range progressCh {
				info.TotalDownloaded += p
				/// info.TotalSize = 0 // TODO: get total size
				resCh <- info
			}
		}()

		for _, file := range files {
			d.downloader.Download(
				ctx,
				fmt.Sprintf("%s/%s/resolve/main/%s", HF_ENDPOINT, modelName, file),
				filepath.Join(outputPath, file),
				progressCh,
			)
		}
	}()
	return resCh, errCh
}
