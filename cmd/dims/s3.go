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

//go:build s3.object.lambda
// +build s3.object.lambda

package main

import (
	"context"
	"github.com/beetlebugorg/go-dims/internal/aws"
	"github.com/beetlebugorg/go-dims/internal/core"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
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
		request, err := aws.NewS3ObjectLambdaRequest(*event, config)
		if err != nil {
			slog.Error("failed to parse request", "error", err)
			return
		}

		if err := dims.Handler(request); err != nil {
			slog.Error("failed to process request", "error", err)

			if err := request.SendError(err); err != nil {
				slog.Error("failed to send error response", "error", err)
			}

			return
		}
	}

	lambda.Start(handler)
}
