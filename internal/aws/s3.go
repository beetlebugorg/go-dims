package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/caarlos0/env/v10"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client *s3.Client

type imageBackend struct {
	Config core.S3
}

func init() {
	envConfig := core.S3{}
	if err := env.Parse(&envConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
	}

	client = s3.NewFromConfig(cfg)

	core.RegisterImageBackend(NewImageBackend(envConfig))
}

func NewImageBackend(config core.S3) core.ImageBackend {
	return imageBackend{
		Config: config,
	}
}

func (backend imageBackend) Name() string {
	return "s3"
}

func (backend imageBackend) CanHandle(imageSource string) bool {
	if strings.HasPrefix(imageSource, "s3://") {
		return true
	}

	return false
}

func (backend imageBackend) FetchImage(imageSource string, timeout time.Duration) (*core.Image, error) {
	slog.Info("downloadImageS3", "url", imageSource)

	bucketName := backend.Config.Bucket
	key := strings.TrimPrefix(imageSource, "/")

	if strings.HasPrefix(imageSource, "s3://") {
		u, err := url.Parse(imageSource)
		if err != nil {
			return nil, err
		}

		bucketName = u.Hostname()
		key = strings.TrimPrefix(u.Path, "/")
	}

	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		slog.Debug("s3.GetObject failed", "bucket", bucketName, "key", key)
		return nil, err
	}

	lastModified := response.LastModified.Format(http.TimeFormat)
	size := int(*response.ContentLength)
	imageBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	sourceImage := core.Image{
		Status:       200,
		Etag:         *response.ETag,
		Size:         size,
		Bytes:        imageBytes,
		Format:       vips.ImageTypes[vips.DetermineImageType(imageBytes)],
		LastModified: lastModified,
	}

	return &sourceImage, nil
}
