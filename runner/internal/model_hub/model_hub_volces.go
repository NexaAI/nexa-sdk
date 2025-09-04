package model_hub

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
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
	key, _ := base64.StdEncoding.DecodeString("QUtMVFptWTNZbUkzTmpJMk1tWmpORGRtTVRsaFkyUmpZVFpoTm1SallXSTVOamM=")
	secret, _ := base64.StdEncoding.DecodeString("VG1wamVVOVhWWGxPTWtwcFQwUnJlVTVIVVhoTlZHaHFXa2RXYVU1SFJUUk9hbFY0V1cxT2EwMXFUUT09")

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
	res, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
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

func (d *Vocles) ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error) {
	modelName = strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/"

	data, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
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

func (d *Vocles) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	name := strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/" + fileName

	slog.Debug("Vocles GetFileContent", "modelName", modelName, "fileName", fileName, "offset", offset, "limit", limit)

	input := &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(name),
	}

	if limit > 0 {
		input.Range = aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1))
	} else if offset > 0 {
		input.Range = aws.String(fmt.Sprintf("bytes=%d-", offset))
	}

	data, err := d.s3Client.GetObject(ctx, input)
	if err != nil {
		return err
	}
	defer data.Body.Close()

	_, err = io.Copy(writer, data.Body)
	return err
}
