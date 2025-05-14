// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

var client *s3.Client

type S3ObjectLambdaRequest struct {
	dims.Request
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
	}

	client = s3.NewFromConfig(cfg)
}

func NewS3ObjectLambdaRequest(event events.S3ObjectLambdaEvent, config core.Config) dims.Request {
	u, err := url.Parse(event.UserRequest.URL)
	if err != nil {
		slog.Error("failed to parse URL", "error", err)
	}

	rawCommands := strings.TrimPrefix(u.Path, "/v5")

	slog.Info("NewS3ObjectRequest", "event", event)
	slog.Info("NewS3ObjectRequest", "URL", event.UserRequest.URL)
	slog.Info("NewS3ObjectRequest", "rawCommands", rawCommands)

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return dims.Request{
		Id:          requestHash,
		Config:      config,
		ClientId:    u.Query().Get("clientId"),
		ImageUrl:    u.Query().Get("url"),
		RawCommands: rawCommands,
		Signature:   u.Query().Get("sig"),
	}
}

// FetchImage downloads the image from the given URL.
func (r *S3ObjectLambdaRequest) FetchImage() error {
	slog.Info("downloadImage", "url", r.ImageUrl)

	image, err := r.fetchImage()
	if err != nil {
		return err
	}

	r.SourceImage = *image

	return nil
}

func (r *S3ObjectLambdaRequest) fetchImage() (*core.Image, error) {
	http.DefaultClient.Timeout = time.Duration(r.Config.Timeout.Download) * time.Millisecond

	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String("shortfinal-go-dims"),
		Key:    aws.String(strings.TrimLeft(r.ImageUrl, "/")),
	})

	if err != nil {
		return nil, err
	}

	sourceImage := core.Image{
		Status: 200,
		Etag:   *response.ETag,
		Format: *response.ContentType,
	}

	sourceImage.Size = int(*response.ContentLength)
	sourceImage.Bytes, _ = io.ReadAll(response.Body)

	return &sourceImage, nil
}
