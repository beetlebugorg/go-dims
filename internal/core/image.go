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

package core

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

type Image struct {
	Bytes        []byte // The downloaded image.
	Size         int    // The original image size in bytes.
	Format       string // The original image format.
	Status       int    // The HTTP status code of the downloaded image.
	CacheControl string // The cache headers from the downloaded image.
	EdgeControl  string // The edge control headers from the downloaded image.
	LastModified string // The last modified header from the downloaded image.
	Etag         string // The etag header from the downloaded image.
}

// ImageTypes defines the various image types supported by govips
var ImageTypes = map[string]vips.ImageType{
	"gif":  vips.ImageTypeGIF,
	"jpeg": vips.ImageTypeJPEG,
	"jpg":  vips.ImageTypeJPEG,
	"png":  vips.ImageTypePNG,
	"tiff": vips.ImageTypeTIFF,
	"webp": vips.ImageTypeWEBP,
	"heif": vips.ImageTypeHEIF,
	"svg":  vips.ImageTypeSVG,
	"psd":  vips.ImageTypePSD,
}

type ImageFetcher interface {
	FetchImage(timeout time.Duration) (*Image, error)
}

func FetchImage(imageUrl string, timeout time.Duration) (*Image, error) {
	slog.Debug("downloadImage", "url", imageUrl)

	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", fmt.Sprintf("go-dims/%s", Version))

	http.DefaultClient.Timeout = timeout
	image, err := http.DefaultClient.Do(request)
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
		Format:       vips.ImageTypes[vips.DetermineImageType(imageBytes)],
		Size:         imageSize,
		Bytes:        imageBytes,
	}

	if image.StatusCode != 200 {
		return nil, &StatusError{
			Message:    fmt.Sprintf("failed to fetch image from %s", imageUrl),
			StatusCode: image.StatusCode,
		}
	}

	return &sourceImage, nil
}

func NewJpegExportParams(options JpegCompression, stripMetadata bool) *vips.JpegExportParams {
	jpegParams := &vips.JpegExportParams{
		StripMetadata:      stripMetadata,
		Quality:            options.Quality,
		Interlace:          options.Interlace,
		OptimizeCoding:     options.OptimizeCoding,
		TrellisQuant:       options.TrellisQuant,
		OvershootDeringing: options.OvershootDeringing,
		OptimizeScans:      options.OptimizeScans,
		QuantTable:         options.QuantTable,
	}

	jpegParams.SubsampleMode = vips.VipsForeignSubsampleOff
	if options.SubsampleMode {
		jpegParams.SubsampleMode = vips.VipsForeignSubsampleAuto
	}

	return jpegParams
}

func NewPngExportParams(options PngCompression, stripMetadata bool) *vips.PngExportParams {
	pngParams := &vips.PngExportParams{
		StripMetadata: stripMetadata,
		Quality:       options.Quality,
		Interlace:     options.Interlace,
	}

	if options.Compression >= 0 && options.Compression <= 9 {
		pngParams.Compression = options.Compression
	}

	return pngParams
}

func NewWebpExportParams(options WebpCompression, stripMetadata bool) *vips.WebpExportParams {
	webpParams := &vips.WebpExportParams{
		StripMetadata:   stripMetadata,
		Quality:         options.Quality,
		ReductionEffort: options.ReductionEffort,
	}

	if options.Compression == "lossless" {
		webpParams.NearLossless = false
		webpParams.Lossless = true
	} else if options.Compression == "near_lossless" {
		webpParams.NearLossless = true
		webpParams.Lossless = true
	} else if options.Compression == "lossy" {
		webpParams.NearLossless = false
		webpParams.Lossless = false
	}

	return webpParams
}
