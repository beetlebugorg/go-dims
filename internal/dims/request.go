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

package dims

import (
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
	"path/filepath"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

type Request struct {
	Id                     string // The hash of the request -> hash(clientId + commands + imageUrl).
	Signature              string // The signature of the request.
	Config                 Config // The global configuration.
	ClientId               string // The client ID of this request.
	ImageUrl               string // The image URL that is being manipulated.
	SendContentDisposition bool   // The content disposition of the request.
	RawCommands            string // The commands ('resize/100x100', 'strip/true/format/png', etc).
	Error                  bool   // Whether the error image is being served.
	SourceImage            Image  // The source image.

	format    *string
	strip     bool
	vipsImage *vips.ImageRef

	// Format information
	exportJpegParams *vips.JpegExportParams
	exportPngParams  *vips.PngExportParams
	exportWebpParams *vips.WebpExportParams
	exportGifParams  *vips.GifExportParams
}

var VipsCommands = map[string]VipsOperation{
	"crop":             CropCommand,
	"resize":           ResizeCommand,
	"strip":            StripMetadataCommand,
	"format":           FormatCommand,
	"quality":          QualityCommand,
	"sharpen":          SharpenCommand,
	"brightness":       BrightnessCommand,
	"flipflop":         FlipFlopCommand,
	"sepia":            SepiaCommand,
	"grayscale":        GrayscaleCommand,
	"autolevel":        AutolevelCommand,
	"invert":           InvertCommand,
	"rotate":           RotateCommand,
	"thumbnail":        ThumbnailCommand,
	"legacy_thumbnail": LegacyThumbnailCommand,
}

// FetchImage downloads the image from the given URL.
func (r *Request) FetchImage() error {
	slog.Info("downloadImage", "url", r.ImageUrl)

	timeout := time.Duration(r.Config.Timeout.Download) * time.Millisecond
	sourceImage, err := _fetchImage(r.ImageUrl, timeout)
	if err != nil {
		return err
	}

	if sourceImage.Status != 200 {
		return fmt.Errorf("failed to download image: %d", sourceImage.Status)
	}

	r.SourceImage = *sourceImage

	return nil
}

func _fetchImage(imageUrl string, timeout time.Duration) (*Image, error) {
	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, err
	}

	http.DefaultClient.Timeout = timeout
	image, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}

	imageSize := int(image.ContentLength)
	imageBytes, err := io.ReadAll(image.Body)
	if err != nil {
		return nil, err
	}

	sourceImage := Image{
		Status:       image.StatusCode,
		EdgeControl:  image.Header.Get("Edge-Control"),
		CacheControl: image.Header.Get("Cache-Control"),
		LastModified: image.Header.Get("Last-Modified"),
		Etag:         image.Header.Get("Etag"),
		Format:       image.Header.Get("Content-Type"),
		Size:         imageSize,
		Bytes:        imageBytes,
	}

	slog.Info("downloadImage", "status", sourceImage.Status, "edgeControl", sourceImage.EdgeControl, "cacheControl", sourceImage.CacheControl, "lastModified", sourceImage.LastModified, "etag", sourceImage.Etag, "format", sourceImage.Format)

	return &sourceImage, nil
}

// ProcessImage will execute the commands on the image.
func (r *Request) ProcessImage() (string, []byte, error) {
	slog.Debug("executeVips")

	image, err := vips.NewImageFromBuffer(r.SourceImage.Bytes)
	if err != nil {
		return "", nil, err
	}
	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	requestedSize, err := r.requestedImageSize()
	if err == nil && vips.DetermineImageType(r.SourceImage.Bytes) == vips.ImageTypeJPEG {
		xs := image.Width() / int(requestedSize.Width)
		ys := image.Height() / int(requestedSize.Height)

		if (xs > 2) || (ys > 2) {
			importParams.JpegShrinkFactor.Set(4)
		}
	}

	image, err = vips.LoadImageFromBuffer(r.SourceImage.Bytes, importParams)
	if err != nil {
		return "", nil, err
	}

	r.vipsImage = image

	slog.Info("executeVips", "image", image, "buffer-size", len(r.SourceImage.Bytes), "strip", r.strip, "format", r.format)

	// Execute the commands.
	stripMetadata := r.Config.StripMetadata

	ctx, task := trace.NewTask(context.Background(), "v5.ProcessImage")
	defer task.End()

	for _, command := range r.Commands() {
		if command.Name == "strip" && command.Args == "true" {
			stripMetadata = false
		}

		region := trace.StartRegion(ctx, command.Name)
		if err := r.ProcessCommand(command); err != nil {
			return "", nil, err
		}
		region.End()
	}

	if stripMetadata {
		// set strip option in export options
	}

	if *r.format == "jpg" {
		imageBytes, _, err := image.ExportJpeg(r.exportJpegParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeJPEG], imageBytes, nil
	} else if *r.format == "png" {
		imageBytes, _, err := image.ExportPng(r.exportPngParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypePNG], imageBytes, nil
	} else if *r.format == "webp" {
		imageBytes, _, err := image.ExportWebp(r.exportWebpParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeWEBP], imageBytes, nil
	} else if *r.format == "gif" {
		imageBytes, _, err := image.ExportGIF(r.exportGifParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[vips.ImageTypeGIF], imageBytes, nil
	}

	imageBytes, _, err := image.ExportNative()
	if err != nil {
		return "", nil, err
	}

	return vips.ImageTypes[image.Format()], imageBytes, nil
}

func (r *Request) SendHeaders(w http.ResponseWriter) {
	maxAge := r.Config.OriginCacheControl.Default
	edgeControlTtl := r.Config.EdgeControl.DownstreamTtl

	if r.Config.OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.Config.OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.Config.OriginCacheControl.Max
			if maxCacheAge != 0 && maxAge >= maxCacheAge {
				maxAge = maxCacheAge
			}
		}
	}

	if r.Error {
		maxAge = r.Config.OriginCacheControl.Error
	}

	// Set cache headers.
	if maxAge > 0 {
		slog.Debug("sendImage", "maxAge", maxAge)

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		w.Header().Set("Expires",
			fmt.Sprintf("%s", time.Now().
				Add(time.Duration(maxAge)*time.Second).
				UTC().
				Format(http.TimeFormat)))
	}

	if edgeControlTtl > 0 {
		w.Header().Set("Edge-Control", fmt.Sprintf("downstream-ttl=%d", edgeControlTtl))
	}

	// Set content disposition.
	if r.SendContentDisposition {
		// Grab filename from imageUrl
		u, err := url.Parse(r.ImageUrl)
		if err != nil {
			return
		}

		filename := filepath.Base(u.Path)

		slog.Debug("sendImage", "sendContentDisposition", r.SendContentDisposition, "filename", filename)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.SourceImage.Etag != "" || r.SourceImage.LastModified != "" {
		var h hash.Hash
		if r.Config.Signing.EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		} else if r.SourceImage.LastModified != "" {
			h.Write([]byte(r.SourceImage.LastModified))
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("Etag", etag)
	}
}

func (r *Request) SendImage(w http.ResponseWriter, status int, imageFormat string, imageBlob []byte) error {
	slog.Info("SendImage", "status", status, "format", imageFormat, "size", len(imageBlob))

	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	w.WriteHeader(status)

	// Write the image.
	_, err := w.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) SendError(w http.ResponseWriter, status int, message string) {
	slog.Info("sendError", "status", status, "message", message)
}

func (r *Request) Commands() []Command {
	commands := make([]Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, Command{
			Name: command,
			Args: args,
		})

		slog.Debug("parsedCommand", "command", command, "args", args)
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

func (r *Request) ProcessCommand(command Command) error {
	if operation, ok := VipsCommands[command.Name]; ok {
		return operation(r, command.Args)
	}

	return fmt.Errorf("command not found: %s", command.Name)
}

// Parse through the requested commands and return requested image size for thumbnail and resize
// operations.
//
// This is used while reading an image to improve performance when generating thumbnails from very
// large images.
func (r *Request) requestedImageSize() (geometry.Geometry, error) {
	for _, command := range r.Commands() {
		if command.Name == "thumbnail" || command.Name == "resize" {
			var rect = geometry.ParseGeometry(command.Args)

			if rect.Width > 0 && rect.Height > 0 {
				return rect, nil
			}

		}
	}

	return geometry.Geometry{}, errors.New("no resize or thumbnail command found")
}
