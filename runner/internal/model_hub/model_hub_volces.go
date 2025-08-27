package model_hub

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/bytedance/sonic"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

type Vocles struct {
	isChinaMainland    bool
	chinaMainlandCheck sync.Once

	s3Client *s3.Client
}

func NewVocles() *Vocles {
	d := &Vocles{}
	d.initS3Client()
	return d
}

var (
	errNotChinaMainland = fmt.Errorf("not in China mainland")
	errNotSupported     = fmt.Errorf("not supported")
)

func (d *Vocles) checkChinaMainland() bool {
	d.chinaMainlandCheck.Do(func() {
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
			d.isChinaMainland = code == "CN"
			return
		}
		slog.Error("Detect country code failed")
	})
	return d.isChinaMainland
}

func (d *Vocles) initS3Client() {
	key, _ := base64.StdEncoding.DecodeString("QUtMVE5qRmxNV001TjJRd1ltVm1OR05qWlRsaVl6ZGlZV0UxTnpBNE4yRm1Zak0=")
	secret, _ := base64.StdEncoding.DecodeString("VDFkU2FsbHRUWHBPVkdjelRrUkplazVIV1ROWlYwbDRUbTFPYVZwVWF6Rk9la1Y2V1RKRmVrOUVUUT09")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(string(key), string(secret), "")),
		config.WithRegion("cn-beijing"),
		config.WithLogger(logging.Nop{}),
	)
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	d.s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://tos-s3-cn-beijing.volces.com")
	})
}

func (d *Vocles) CheckAvailable(ctx context.Context, modelName string) error {
	if !d.checkChinaMainland() {
		return errNotChinaMainland
	}

	if !strings.HasPrefix(modelName, "NexaAI/") {
		return errNotSupported
	}

	modelName = strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/"
	_, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String("nexa-model-hub-bucket"),
		Prefix:  aws.String(modelName),
		MaxKeys: aws.Int32(1),
	})

	return err
}

func (d *Vocles) ModelInfo(ctx context.Context, modelName string) ([]string, error) {
	modelName = strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/"

	data, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Prefix: aws.String(modelName),
	})
	if err != nil {
		return nil, err
	}

	res := make([]string, len(data.Contents))
	for i, item := range data.Contents {
		res[i] = strings.TrimPrefix(*item.Key, modelName)
	}
	return res, nil
}

func (d *Vocles) FileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	name := strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/" + fileName

	data, err := d.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(name),
	})

	if err != nil {
		return 0, err
	}

	return *data.ContentLength, nil
}

func (d *Vocles) GetQuantInfo(ctx context.Context, modelName string) (int, error) {
	return -1, errNotSupported
}

func (d *Vocles) StartDownload(ctx context.Context, modelName, outputPath string, files []string) (chan types.DownloadInfo, chan error) {
	slog.Debug("Vocles StartDownload", "modelName", modelName, "outputPath", outputPath, "files", files)

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
			input := strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/" + file
			if err := d.downloadFile(ctx, input, filepath.Join(outputPath, file), progressCh); err != nil {
				errCh <- err
				return
			}
		}
	}()
	return resCh, errCh
}

func (d *Vocles) downloadFile(ctx context.Context, input, output string, progress chan int64) error {
	slog.Debug("Vocles Download", "input", input, "output", output)

	// download from volces
	data, err := d.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(input),
	})

	if err != nil {
		return err
	}
	defer data.Body.Close()

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = outFile.Truncate(*data.ContentLength)
	if err != nil {
		return err
	}

	for {
		n, err := io.CopyN(outFile, data.Body, 4*1024*1024) // update progress every 4MB

		if err == io.EOF {
			progress <- n
			return nil
		}

		if err != nil {
			return err
		}

		progress <- n
	}
}
