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

package main

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/dims/aws"
	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/davidbyttow/govips/v2/vips"
)

type AwsCmd struct {
	S3Object S3ObjectCmd `cmd:"" name:"s3-lambda" help:"Implementation of AWS S3 Object Lambda."`
}

type S3ObjectCmd struct {
}

func (s *S3ObjectCmd) Run() error {
	config := core.ReadConfig()

	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)

	var opts *slog.HandlerOptions
	if config.DebugMode {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	handler := func(ctx context.Context, event *events.S3ObjectLambdaEvent) {
		request := aws.NewS3ObjectLambdaRequest(*event, config)
		response := httptest.NewRecorder()

		dims.Handler(request, config, response)

		httpResponse := response.Result()

		cfg, err := awscfg.LoadDefaultConfig(ctx)
		if err != nil {
			slog.Error("failed to load configuration", "error", err)
			return
		}
		cfg.Region = os.Getenv("AWS_REGION")
		svc := s3.NewFromConfig(cfg)

		statusCode := int32(httpResponse.StatusCode)
		etag := httpResponse.Header.Get("ETag")
		contentType := httpResponse.Header.Get("Content-Type")
		cacheControl := httpResponse.Header.Get("Cache-Control")
		cl, err := strconv.Atoi(httpResponse.Header.Get("Content-Length"))
		if err != nil {
			slog.Error("failed to parse Content-Length", "error", err)
			return
		}
		contentLength := int64(cl)

		if _, err := svc.WriteGetObjectResponse(ctx, &s3.WriteGetObjectResponseInput{
			StatusCode:    &statusCode,
			ETag:          &etag,
			ContentType:   &contentType,
			ContentLength: &contentLength,
			CacheControl:  &cacheControl,
			Body:          httpResponse.Body,
			RequestRoute:  &event.GetObjectContext.OutputRoute,
			RequestToken:  &event.GetObjectContext.OutputToken,
		}); err != nil {
			slog.Error("failed to write get object response", "error", err)
			return
		}
	}

	lambda.Start(handler)

	return nil
}
