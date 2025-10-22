package model_hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/bytedance/sonic"
	"resty.dev/v3"
)

type Volces struct {
	isChinaMainland    bool
	chinaMainlandCheck sync.Once

	s3Client *s3.Client
}

func NewVolces(skipCNCheck bool) *Volces {
	v := &Volces{}
	if skipCNCheck {
		v.chinaMainlandCheck.Do(func() {
			v.isChinaMainland = true
			slog.Info("Skip China mainland check for Volces model hub")
		})
	}
	v.initS3Client()
	return v
}

var (
	errNotChinaMainland = fmt.Errorf("not in China mainland")
	errNotSupported     = fmt.Errorf("not supported")
)

func (v *Volces) checkChinaMainland() bool {
	v.chinaMainlandCheck.Do(func() {
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
			v.isChinaMainland = code == "CN"
			return
		}
		slog.Error("Detect country code failed")
	})
	return v.isChinaMainland
}

func (v *Volces) initS3Client() {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithLogger(logging.Nop{}),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithBaseEndpoint("https://tos-s3-cn-beijing.volces.com"),
		config.WithRegion("cn-beijing"),
	)
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	v.s3Client = s3.NewFromConfig(cfg)
}

func (v *Volces) CheckAvailable(ctx context.Context, modelName string) error {
	if !v.checkChinaMainland() {
		return errNotChinaMainland
	}

	if !strings.HasPrefix(modelName, "NexaAI/") {
		return errNotSupported
	}

	modelName = strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/"
	res, err := v.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String("nexa-model-hub-bucket"),
		Prefix:  aws.String(modelName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return err
	}

	if aws.ToInt32(res.KeyCount) == 0 {
		return fmt.Errorf("model %s not found", modelName)
	}

	return nil
}

func (v *Volces) MaxConcurrency() int {
	return 4
}

func (v *Volces) ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error) {
	modelName = strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/"

	data, err := v.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Prefix: aws.String(modelName),
	})
	if err != nil {
		return nil, err
	}

	res := make([]ModelFileInfo, len(data.Contents))
	for i, item := range data.Contents {
		res[i] = ModelFileInfo{
			Name: strings.TrimPrefix(*item.Key, modelName),
			Size: *item.Size,
		}
	}
	return res, nil
}

func (v *Volces) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	name := strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/" + fileName

	slog.Debug("Volces GetFileContent", "modelName", modelName, "fileName", fileName, "name", name, "offset", offset, "limit", limit)

	input := &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(name),
	}

	if limit > 0 {
		input.Range = aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1))
	} else if offset > 0 {
		input.Range = aws.String(fmt.Sprintf("bytes=%d-", offset))
	}

	data, err := v.s3Client.GetObject(ctx, input)
	if err != nil {
		return err
	}
	defer data.Body.Close()

	_, err = io.Copy(writer, data.Body)
	return err
}
