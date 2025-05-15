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

package request

import (
	"context"
	"errors"
	"fmt"
	core2 "github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/geometry"
	operations2 "github.com/beetlebugorg/go-dims/internal/operations"
	"net/http"
	"net/url"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

type DimsRequest struct {
	Id                     string       // The hash of the request -> hash(clientId + commands + imageUrl).
	URL                    *url.URL     // The URL of the request.
	ImageUrl               string       // The image URL that is being manipulated.
	SendContentDisposition bool         // The content disposition of the request.
	RawCommands            string       // The commands ('resize/100x100', 'strip/true/format/png', etc).
	SourceImage            core2.Image  // The source image.
	config                 core2.Config // The global configuration.
	shrinkFactor           int
}

type HttpDimsRequest struct {
	DimsRequest

	request  http.Request
	response http.ResponseWriter
}

var VipsTransformCommands = map[string]operations2.VipsTransformOperation{
	"crop":             operations2.CropCommand,
	"resize":           operations2.ResizeCommand,
	"sharpen":          operations2.SharpenCommand,
	"brightness":       operations2.BrightnessCommand,
	"flipflop":         operations2.FlipFlopCommand,
	"sepia":            operations2.SepiaCommand,
	"grayscale":        operations2.GrayscaleCommand,
	"autolevel":        operations2.AutolevelCommand,
	"invert":           operations2.InvertCommand,
	"rotate":           operations2.RotateCommand,
	"thumbnail":        operations2.ThumbnailCommand,
	"legacy_thumbnail": operations2.LegacyThumbnailCommand,
}

var VipsExportCommands = map[string]operations2.VipsExportOperation{
	"strip":   operations2.StripMetadataCommand,
	"format":  operations2.FormatCommand,
	"quality": operations2.QualityCommand,
}

var VipsRequestCommands = map[string]operations2.VipsRequestOperation{
	"watermark": operations2.Watermark,
}

//-- Request/RequestContext Implementation

func NewHttpDimsRequest(r http.Request, w http.ResponseWriter, id string, imageUrl string, commands string, config core2.Config) *HttpDimsRequest {
	return &HttpDimsRequest{
		DimsRequest: *NewDimsRequest(id, r.URL, imageUrl, commands, config),
		request:     r,
		response:    w,
	}
}

func NewDimsRequest(id string, url *url.URL, imageUrl string, commands string, config core2.Config) *DimsRequest {
	return &DimsRequest{
		Id:          id,
		URL:         url,
		ImageUrl:    imageUrl,
		RawCommands: commands,
		config:      config,
	}
}

func (r *DimsRequest) Config() core2.Config {
	return r.config
}

func (r *DimsRequest) LoadImage(sourceImage *core2.Image) (*vips.ImageRef, error) {
	image, err := vips.NewImageFromBuffer(sourceImage.Bytes)
	if err != nil {
		return nil, err
	}
	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	r.shrinkFactor = 1
	requestedSize, err := r.requestedImageSize()
	if err == nil && vips.DetermineImageType(sourceImage.Bytes) == vips.ImageTypeJPEG {
		xs := image.Width() / int(requestedSize.Width)
		ys := image.Height() / int(requestedSize.Height)

		if (xs > 2) || (ys > 2) {
			importParams.JpegShrinkFactor.Set(4)
			r.shrinkFactor = 4
		}
	}

	return vips.LoadImageFromBuffer(sourceImage.Bytes, importParams)
}

// ProcessImage will execute the commands on the image.
func (r *DimsRequest) ProcessImage(image *vips.ImageRef, errorImage bool) (string, []byte, error) {
	ctx := context.Background()

	// Execute the commands.
	ctx, task := trace.NewTask(ctx, "v5.ProcessImage")
	defer task.End()

	opts := operations2.ExportOptions{
		ImageType:        core2.ImageTypes[r.SourceImage.Format],
		JpegExportParams: core2.NewJpegExportParams(r.config.ImageOutputOptions.Jpeg, r.config.StripMetadata),
		PngExportParams:  core2.NewPngExportParams(r.config.ImageOutputOptions.Png, r.config.StripMetadata),
		WebpExportParams: core2.NewWebpExportParams(r.config.ImageOutputOptions.Webp, r.config.StripMetadata),
		GifExportParams:  vips.NewGifExportParams(),
		TiffExportParams: vips.NewTiffExportParams(),
	}

	stripMetadata := r.config.StripMetadata
	opts.GifExportParams.StripMetadata = stripMetadata
	opts.TiffExportParams.StripMetadata = stripMetadata

	for _, command := range r.Commands() {
		region := trace.StartRegion(ctx, command.Name)

		if operation, ok := VipsTransformCommands[command.Name]; ok {
			if command.Name == "crop" {
				adjustedArgs, err := adjustCropAfterShrink(command.Args, r.shrinkFactor)
				if err != nil {
					return "", nil, err
				}

				command.Args = adjustedArgs
			}

			if command.Name == "strip" && command.Args != "true" {
				stripMetadata = false
			}

			if err := operation(image, command.Args); err != nil && !errorImage {
				return "", nil, err
			}
		} else if operation, ok := VipsExportCommands[command.Name]; ok {
			if err := operation(image, command.Args, &opts); err != nil && !errorImage {
				return "", nil, err
			}
		} else if operation, ok := VipsRequestCommands[command.Name]; ok && !errorImage {
			if err := operation(image, command.Args, operations2.RequestOperation{
				Config: r.config,
				URL:    r.URL,
			}); err != nil {
				return "", nil, err
			}
		}

		region.End()
	}

	if stripMetadata {
		image.RemoveMetadata()
	}

	switch opts.ImageType {
	case vips.ImageTypeJPEG:
		imageBytes, _, err := image.ExportJpeg(opts.JpegExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeJPEG], imageBytes, nil

	case vips.ImageTypePNG:
		imageBytes, _, err := image.ExportPng(opts.PngExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypePNG], imageBytes, nil

	case vips.ImageTypeWEBP:
		imageBytes, _, err := image.ExportWebp(opts.WebpExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeWEBP], imageBytes, nil
	case vips.ImageTypeGIF:
		imageBytes, _, err := image.ExportGIF(opts.GifExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeGIF], imageBytes, nil
	case vips.ImageTypeTIFF:
		imageBytes, _, err := image.ExportTiff(opts.TiffExportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeTIFF], imageBytes, nil
	}

	imageBytes, _, err := image.ExportNative()
	if err != nil {
		return "", nil, err
	}

	return vips.ImageTypes[opts.ImageType], imageBytes, nil
}

func (r *DimsRequest) FetchImage(timeout time.Duration) (*core2.Image, error) {
	image, err := core2.FetchImage(r.ImageUrl, timeout)
	if err != nil {
		return nil, err
	}

	r.SourceImage = *image

	return image, nil
}

func (r *DimsRequest) Commands() []operations2.Command {
	commands := make([]operations2.Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, operations2.Command{
			Name: command,
			Args: args,
		})
	}

	return commands
}

func sourceMaxAge(header string) (int, error) {
	if header == "" {
		return 0, errors.New("empty header")
	}

	pattern, _ := regexp.Compile(`max-age=(\d+)`)
	match := pattern.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0, errors.New("unable to convert to int")
		}

		return sourceMaxAge, nil
	}

	return 0, errors.New("max-age not found in header")
}

// Parse through the requested commands and return requested image size for thumbnail and resize
// operations.
//
// This is used while reading an image to improve performance when generating thumbnails from very
// large images.
func (r *DimsRequest) requestedImageSize() (geometry.Geometry, error) {
	for _, command := range r.Commands() {
		if command.Name == "thumbnail" || command.Name == "resize" {
			rect, err := geometry.ParseGeometry(command.Args)
			if err != nil {
				return geometry.Geometry{}, err
			}

			if rect.Width > 0 && rect.Height > 0 {
				return rect, nil
			}

		}
	}

	return geometry.Geometry{}, errors.New("no resize or thumbnail command found")
}

func adjustCropAfterShrink(args string, factor int) (string, error) {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return "", operations2.NewOperationError("crop", args, err.Error())
	}

	rect.X = int(float64(rect.X) / float64(factor))
	rect.Y = int(float64(rect.Y) / float64(factor))

	if rect.Width > 0 {
		rect.Width = float64(rect.Width) / float64(factor)
	}

	if rect.Height > 0 {
		rect.Height = float64(rect.Height) / float64(factor)
	}

	// Output full geometry
	if rect.Y > 0 {
		return fmt.Sprintf("%dx%d+%d+%d", int(rect.Width), int(rect.Height), int(rect.X), int(rect.Y)), nil
	}

	// Output geometry without Y
	if rect.X > 0 {
		return fmt.Sprintf("%dx%d+%d", int(rect.Width), int(rect.Height), int(rect.X)), nil
	}

	// Output geometry without offsets
	if rect.Height > 0 {
		return fmt.Sprintf("%dx%d", int(rect.Width), int(rect.Height)), nil
	}

	return fmt.Sprintf("%dx", int(rect.Width)), nil
}
