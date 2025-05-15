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
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/beetlebugorg/go-dims/internal/dims/operations"
	"github.com/beetlebugorg/go-dims/internal/dims/request"
	v4 "github.com/beetlebugorg/go-dims/internal/dims/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/dims/v5"
	"github.com/beetlebugorg/go-dims/internal/gox/imagex/colorx"
	"github.com/davidbyttow/govips/v2/vips"
)

var client *s3.Client

type S3ObjectLambdaRequest struct {
	request.DimsRequest
	event events.S3ObjectLambdaEvent
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

		if !config.DevelopmentMode &&
			!v5.ValidateSignatureV5(rawCommands, u.Query().Get("url"), signedKeys, config.SigningKey, signature) {
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

		if !config.DevelopmentMode &&
			!v4.ValidateSignatureV4(rawCommands, timestamp, u.Query().Get("url"), signedKeys, config.SigningKey, signature) {
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

	return &S3ObjectLambdaRequest{
		DimsRequest: *request.NewDimsRequest(requestHash, u, u.Query().Get("url"), rawCommands, config),
		event:       event,
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

	r.SourceImage = sourceImage

	return &sourceImage, nil
}

func (r *S3ObjectLambdaRequest) SendError(err error) error {
	message := err.Error()

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var statusError *core.StatusError
	var operationError *operations.OperationError
	if errors.As(err, &statusError) {
		status = statusError.StatusCode
	} else if errors.As(err, &operationError) {
		status = operationError.StatusCode
	}

	errorImage, err := vips.Black(512, 512)
	if err != nil {
		return err
	}

	if err := errorImage.BandJoinConst([]float64{0, 0}); err != nil {
		return err
	}

	backgroundColor, err := colorx.ParseHexColor(r.Config().Error.Background)
	if err != nil {
		return err
	}

	red, green, blue, _ := backgroundColor.RGBA()
	redI := float64(red) / 65535 * 255
	greenI := float64(green) / 65535 * 255
	blueI := float64(blue) / 65535 * 255

	if err := errorImage.Linear([]float64{0, 0, 0}, []float64{redI, greenI, blueI}); err != nil {
		return err
	}

	r.SourceImage = core.Image{
		Status: status,
		Format: vips.ImageTypes[vips.ImageTypeJPEG],
	}

	// Send error headers.
	maxAge := r.Config().OriginCacheControl.Error
	if maxAge > 0 {
		//r.response.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		//r.response.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
	}

	imageType, imageBlob, err := r.ProcessImage(errorImage, true)
	if err != nil {
		// If processing failed because of a bad command then return the image as-is.
		exportOptions := vips.NewJpegExportParams()
		exportOptions.Quality = 1
		imageBytes, _, _ := errorImage.ExportJpeg(exportOptions)

		return r.SendImage(status, "jpg", imageBytes)
	}

	if imageType == "" {
		imageType = "jpg"
	}

	return r.SendImage(status, imageType, imageBlob)
}

func (r *S3ObjectLambdaRequest) SendImage(status int, imageFormat string, imageBlob []byte) error {
	slog.Info("sendImageS3", "url", r.ImageUrl)

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}
	cfg.Region = os.Getenv("AWS_REGION")
	svc := s3.NewFromConfig(cfg)

	var etag string
	// Set etag header.
	if r.SourceImage.Etag != "" {
		var h hash.Hash
		if r.Config().EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		}

		etag = fmt.Sprintf("%x", h.Sum(nil))
	}

	statusCode := int32(status)
	contentType := "image/" + strings.ToLower(imageFormat)
	//cacheControl := httpResponse.Header.Get("Cache-Control")
	contentLength := int64(len(imageBlob))

	if _, err := svc.WriteGetObjectResponse(ctx, &s3.WriteGetObjectResponseInput{
		StatusCode:    &statusCode,
		ETag:          &etag,
		ContentType:   &contentType,
		ContentLength: &contentLength,
		//CacheControl:  &cacheControl,
		Body:         bytes.NewReader(imageBlob),
		RequestRoute: &r.event.GetObjectContext.OutputRoute,
		RequestToken: &r.event.GetObjectContext.OutputToken,
	}); err != nil {
		slog.Error("failed to write get object response", "error", err)
		return err
	}

	return nil
}
