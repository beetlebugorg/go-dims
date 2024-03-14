package dims

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/lambda"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
	"net/http/httptest"
	"os"
	"strconv"
)

func NewLambdaFunctionURLHandler(devMode bool, debugMode bool) func(ctx context.Context, event *events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLResponse, error) {
	imagick.Initialize()

	environmentConfig := dims.ReadConfig()

	config := dims.Config{
		EnvironmentConfig: environmentConfig,
		DevelopmentMode:   devMode,
		DebugMode:         debugMode,
		EtagAlgorithm:     "hmac-sha256",
	}

	return func(ctx context.Context, event *events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLResponse, error) {
		return handleLambdaFunctionURLRequest(ctx, event, config)
	}
}

func handleLambdaFunctionURLRequest(_ context.Context, event *events.LambdaFunctionURLRequest, config dims.Config) (*events.LambdaFunctionURLResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	request := lambda.NewLambdaFunctionURLRequest(*event, config)
	response := httptest.NewRecorder()
	dims.Handler(request, config, response)

	httpResponse := response.Result()

	dimsResponse := &events.LambdaFunctionURLResponse{}
	dimsResponse.StatusCode = httpResponse.StatusCode

	if response.Body.Bytes() != nil {
		dimsResponse.Headers = map[string]string{
			"Content-Type":   httpResponse.Header.Get("Content-Type"),
			"Content-Length": httpResponse.Header.Get("Content-Length"),
		}
		dimsResponse.IsBase64Encoded = true
		dimsResponse.Body = base64.StdEncoding.EncodeToString(response.Body.Bytes())
	} else {
		dimsResponse.Headers = map[string]string{
			"Content-Type": "text/plain",
		}
		dimsResponse.IsBase64Encoded = false
		dimsResponse.Body = "fail"
	}

	slog.Info("HandleRequest", "response", dimsResponse, "event", event)

	return dimsResponse, nil
}

func NewLambdaS3ObjectHandler(devMode bool, debugMode bool) func(ctx context.Context, event *events.S3ObjectLambdaEvent) error {
	imagick.Initialize()

	environmentConfig := dims.ReadConfig()

	config := dims.Config{
		EnvironmentConfig: environmentConfig,
		DevelopmentMode:   devMode,
		DebugMode:         debugMode,
		EtagAlgorithm:     "hmac-sha256",
	}

	return func(ctx context.Context, event *events.S3ObjectLambdaEvent) error {
		return handleLambdaS3ObjectRequest(ctx, event, config)
	}
}

func handleLambdaS3ObjectRequest(ctx context.Context, event *events.S3ObjectLambdaEvent, config dims.Config) error {
	if event == nil {
		return fmt.Errorf("received nil event")
	}

	request := lambda.NewS3ObjectRequest(*event, config)
	response := httptest.NewRecorder()
	dims.Handler(request, config, response)

	httpResponse := response.Result()

	cfg, err := awscfg.LoadDefaultConfig(ctx)
	cfg.Region = os.Getenv("AWS_REGION")
	svc := s3.NewFromConfig(cfg)

	statusCode := int32(httpResponse.StatusCode)
	etag := httpResponse.Header.Get("ETag")
	contentType := httpResponse.Header.Get("Content-Type")
	cl, err := strconv.Atoi(httpResponse.Header.Get("Content-Length"))
	if err != nil {

	}
	contentLength := int64(cl)

	if _, err := svc.WriteGetObjectResponse(ctx, &s3.WriteGetObjectResponseInput{
		StatusCode:    &statusCode,
		ETag:          &etag,
		ContentType:   &contentType,
		ContentLength: &contentLength,
		Body:          httpResponse.Body,
		RequestRoute:  &event.GetObjectContext.OutputRoute,
		RequestToken:  &event.GetObjectContext.OutputToken,
	}); err != nil {
		return err
	}

	return nil
}
