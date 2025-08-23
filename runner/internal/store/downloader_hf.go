package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
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
func (s *Store) HFGetQuantInfo(ctx context.Context, modelName string) (int, error) {
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

func httpCodeToError(c *resty.Client, r *resty.Response) error {
	if r.StatusCode() >= 400 {
		return fmt.Errorf("HTTPError: [%d] %s", r.StatusCode(), r.String())
	}
	return nil
}
