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
	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/beetlebugorg/go-dims/internal/dims/request"
	v4 "github.com/beetlebugorg/go-dims/internal/dims/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/dims/v5"
)

var client *s3.Client

type S3ObjectLambdaRequest struct {
	request.Request
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

	var signature string
	var rawCommands string

	// Signed Parameters
	// _keys query parameter is a comma delimted list of keys to include in the signature.
	var signedKeys []string
	params := u.Query().Get("_keys")
	if params != "" {
		keys := strings.Split(params, ",")
		for _, key := range keys {
			value := u.Query().Get(key)
			if value != "" {
				signedKeys = append(signedKeys, value)
			}
		}
	}

	if strings.HasPrefix(u.Path, "/v5/") {
		rawCommands = strings.TrimPrefix(u.Path, "/v5/")
		signature = u.Query().Get("sig")

		if !v5.ValidateSignatureV5(rawCommands, u.Query().Get("url"), signedKeys, config.SigningKey, signature) {
			return nil, &core.StatusError{
				StatusCode: http.StatusUnauthorized,
				Message:    "invalid signature",
			}
		}
	} else if strings.HasPrefix(u.Path, "/dims4/") {
		// Remove the "/dims4/" prefix if it exists
		// Parse out /<clientId>/<sig>/<expire>/<rawCommands>
		u.Path = strings.TrimPrefix(u.Path, "/dims4/")

		parts := strings.SplitN(u.Path, "/", 4)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid dims4 path format")
		}

		signature = parts[1]
		timestamp := parts[2]
		rawCommands = strings.Join(parts[3:], "/")

		if !v4.ValidateSignatureV4(rawCommands, timestamp, u.Query().Get("url"), signedKeys, config.SigningKey, signature) {
			return nil, &core.StatusError{
				StatusCode: http.StatusUnauthorized,
				Message:    "invalid signature",
			}
		}
	}

	slog.Info("NewS3ObjectRequest", "event", event)
	slog.Info("NewS3ObjectRequest", "URL", event.UserRequest.URL)
	slog.Info("NewS3ObjectRequest", "rawCommands", rawCommands)

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	httpRequest := &http.Request{
		URL: u,
	}

	return &S3ObjectLambdaRequest{
		Request: *request.NewDimsRequest(*httpRequest, requestHash, u.Query().Get("url"), rawCommands, config),
	}, nil
}

func (r *S3ObjectLambdaRequest) FetchImage(timeout time.Duration) (*core.Image, error) {
	slog.Info("downloadImageS3", "url", r.ImageUrl)

	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r.Config().S3.Bucket),
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
