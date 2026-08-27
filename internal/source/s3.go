package source

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
	"sync"
	"time"
)

type s3SourceBackend struct {
	Config core.S3
}

// s3Client is built on first use. Loading AWS configuration at startup probes
// the instance metadata endpoint on a host that has no credentials, which
// slows every start whether or not the backend is ever used.
var s3Client = sync.OnceValues(func() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
})

func init() {
	envConfig := core.S3{}
	if err := env.Parse(&envConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}

	core.RegisterImageBackend(NewS3SourceBackend(envConfig))
}

func NewS3SourceBackend(config core.S3) core.SourceBackend {
	return s3SourceBackend{
		Config: config,
	}
}

func (backend s3SourceBackend) Name() string {
	return "s3"
}

func (backend s3SourceBackend) CanHandle(imageSource string) bool {
	if strings.HasPrefix(imageSource, "s3://") {
		return true
	}

	return false
}

func (backend s3SourceBackend) FetchImage(imageSource string, timeout time.Duration) (*core.Image, error) {
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

	client, err := s3Client()
	if err != nil {
		slog.Error("failed to load AWS configuration", "error", err)
		return nil, core.NewStatusError(500, "s3 backend is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		slog.Debug("s3.GetObject failed", "bucket", bucketName, "key", key)
		return nil, err
	}
	defer response.Body.Close()

	imageBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageFromResponse(response, imageBytes), nil
}

// imageFromResponse maps an S3 response onto an Image. Every field it reads
// is optional on GetObjectOutput, so each one goes through a nil check. A
// response that omits one used to panic the request goroutine.
func imageFromResponse(response *s3.GetObjectOutput, imageBytes []byte) *core.Image {
	var lastModified string
	if response.LastModified != nil {
		lastModified = response.LastModified.UTC().Format(http.TimeFormat)
	}

	return &core.Image{
		Status:       200,
		Etag:         aws.ToString(response.ETag),
		Size:         int(aws.ToInt64(response.ContentLength)),
		Bytes:        imageBytes,
		Format:       vips.DetermineImageType(imageBytes),
		LastModified: lastModified,
	}
}
