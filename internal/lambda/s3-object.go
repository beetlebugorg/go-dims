// Copyright 2024 Jeremy Collins. All rights reserved.
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

package lambda

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RequestLambdaS3Object struct {
	v5.RequestV5
}

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
	}

	client = s3.NewFromConfig(cfg)
}

func NewS3ObjectRequest(event events.S3ObjectLambdaEvent, config dims.Config) *RequestLambdaS3Object {
	u, err := url.Parse(event.UserRequest.URL)
	if err != nil {
		return nil
	}

	rawCommands := strings.TrimPrefix(u.Path, "/v5/dims/")

	slog.Info("NewS3ObjectRequest", "event", event)
	slog.Info("NewS3ObjectRequest", "URL", event.UserRequest.URL)
	slog.Info("NewS3ObjectRequest", "rawCommands", rawCommands)

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestLambdaS3Object{
		RequestV5: v5.RequestV5{
			Request: dims.Request{
				Id:          requestHash,
				Config:      config,
				ClientId:    u.Query().Get("clientId"),
				ImageUrl:    u.Query().Get("url"),
				RawCommands: rawCommands,
				Signature:   u.Query().Get("sig"),
			},
		},
	}
}

// FetchImage downloads the image from the given URL.
func (r *RequestLambdaS3Object) FetchImage() error {
	slog.Info("downloadImage", "url", r.ImageUrl)

	timeout := time.Duration(r.Config.Timeout.Download) * time.Millisecond
	image, err := _fetchImage(r.Request.ImageUrl, timeout)
	if err != nil {
		return err
	}

	r.SourceImage = *image

	return nil
}

func _fetchImage(imageUrl string, timeout time.Duration) (*dims.Image, error) {
	http.DefaultClient.Timeout = timeout

	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String("shortfinal-go-dims"),
		Key:    aws.String(strings.TrimLeft(imageUrl, "/")),
	})

	if err != nil {
		return nil, err
	}

	sourceImage := dims.Image{
		Status: 200,
		Etag:   *response.ETag,
		Format: *response.ContentType,
	}

	sourceImage.Size = int(*response.ContentLength)
	sourceImage.Bytes, _ = io.ReadAll(response.Body)

	return &sourceImage, nil
}
