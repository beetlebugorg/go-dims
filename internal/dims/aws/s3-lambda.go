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

func NewS3ObjectLambdaRequest(event events.S3ObjectLambdaEvent, config core.Config) (*S3ObjectLambdaRequest, error) {
	u, err := url.Parse(event.UserRequest.URL)
	if err != nil {
		slog.Error("failed to parse URL", "error", err)
		return nil, err
	}

	var clientId string
	var signature string
	var rawCommands string

	if strings.HasPrefix(u.Path, "/v5/") {
		rawCommands = strings.TrimPrefix(u.Path, "/v5/")
		signature = u.Query().Get("sig")
		clientId = ""
	} else if strings.HasPrefix(u.Path, "/dims4/") {
		// Remove the "/dims4/" prefix if it exists
		// Parse out /<clientId>/<sig>/<expire>/<rawCommands>
		u.Path = strings.TrimPrefix(u.Path, "/dims4/")

		parts := strings.SplitN(u.Path, "/", 4)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid dims4 path format")
		}

		clientId = parts[0]
		signature = parts[1]
		rawCommands = strings.Join(parts[3:], "/")
	}

	slog.Info("NewS3ObjectRequest", "event", event)
	slog.Info("NewS3ObjectRequest", "URL", event.UserRequest.URL)
	slog.Info("NewS3ObjectRequest", "rawCommands", rawCommands)

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &S3ObjectLambdaRequest{
		dims.Request{
			Id:          requestHash,
			Config:      config,
			ClientId:    clientId,
			ImageUrl:    u.Query().Get("url"),
			RawCommands: rawCommands,
			Signature:   signature,
		},
	}, nil
}

func (r *S3ObjectLambdaRequest) FetchImage(timeout time.Duration) (*core.Image, error) {
	slog.Info("downloadImageS3", "url", r.ImageUrl)

	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r.Config.S3.Bucket),
		Key:    aws.String(strings.TrimLeft(r.ImageUrl, "/")),
	})

	if err != nil {
		return nil, err
	}

	slog.Debug("fetchImage", "response", response)

	lastModified := response.LastModified.Format(http.TimeFormat)

	sourceImage := core.Image{
		Status:       200,
		Etag:         *response.ETag,
		Format:       *response.ContentType,
		LastModified: lastModified,
	}

	sourceImage.Size = int(*response.ContentLength)
	sourceImage.Bytes, _ = io.ReadAll(response.Body)

	return &sourceImage, nil
}
