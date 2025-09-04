package store

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

var vcClient = func() *s3.Client {
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
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://tos-s3-cn-beijing.volces.com")
	})

	return client
}()

func (s *Store) VCModelInfo(ctx context.Context, name string) ([]string, error) {
	name = strings.ReplaceAll(name, "NexaAI/", "model/") + "/"

	data, err := vcClient.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Prefix: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	res := make([]string, len(data.Contents))
	for i, item := range data.Contents {
		res[i] = strings.TrimPrefix(*item.Key, name)
	}
	return res, nil
}

func (s *Store) VCFileSize(ctx context.Context, modelName, fileName string) (int64, error) {
	name := strings.ReplaceAll(modelName, "NexaAI/", "model/") + "/" + fileName

	data, err := vcClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(name),
	})

	if err != nil {
		return 0, err
	}

	return *data.ContentLength, nil
}

func (s *Store) VCGetQuantInfo(ctx context.Context, modelName string) (int, error) {
	return -1, fmt.Errorf("not implemented")
}

func NewVCDownloader(totalSize int64, progress chan<- types.DownloadInfo) Downloader {
	return &VCDownloader{
		totalSize: totalSize,
		progress:  progress,
	}
}

type VCDownloader struct {
	filename        string
	totalSize       int64
	totalDownloaded atomic.Int64
	downloaded      atomic.Int64
	progress        chan<- types.DownloadInfo
}

func (d *VCDownloader) Download(ctx context.Context, url, outputPath string) error {
	slog.Debug("VC Download", "url", url, "outputPath", outputPath)

	// parse url
	url = strings.TrimPrefix(url, HF_ENDPOINT+"/")
	url = strings.TrimSuffix(url, "?download=true")
	url = strings.ReplaceAll(url, "/resolve/main/", "/")

	name := strings.ReplaceAll(url, "NexaAI/", "model/")
	slog.Info("VC Download", "objectKey", name)
	d.filename = filepath.Base(name)

	// download from volces

	data, err := vcClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("nexa-model-hub-bucket"),
		Key:    aws.String(name),
	})

	if err != nil {
		return err
	}
	defer data.Body.Close()

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	err = file.Truncate(*data.ContentLength)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %v", err)
	}

	for {
		n, err := io.CopyN(file, data.Body, 4*1024*1024)
		if err != nil {
			if err == io.EOF {
				d.downloaded.Add(n)
				d.totalDownloaded.Add(n)
				d.progress <- types.DownloadInfo{
					CurrentName:       d.filename,
					CurrentSize:       d.totalSize,
					CurrentDownloaded: d.downloaded.Load(),
					TotalSize:         d.totalSize,
					TotalDownloaded:   d.totalDownloaded.Load(),
				}
				break
			}
			return err
		}
		d.downloaded.Add(n)
		d.totalDownloaded.Add(n)
		d.progress <- types.DownloadInfo{
			CurrentName:       d.filename,
			CurrentSize:       d.totalSize,
			CurrentDownloaded: d.downloaded.Load(),
			TotalSize:         d.totalSize,
			TotalDownloaded:   d.totalDownloaded.Load(),
		}
	}

	return nil
}
