//go:build lambda

package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beetlebugorg/go-dims/internal/aws"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"os"
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

	handler := func(ctx context.Context, event *events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
		request, err := aws.NewRequest(*event, *config)
		if err != nil {
			return nil, err
		}

		if err := dims.Handler(request); err != nil {
			if err := request.SendError(err); err != nil {
				return nil, err
			}
		}

		return request.Response(), nil
	}

	lambda.Start(handler)
}
