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

type ImageBackend interface {
	Name() string
	CanHandle(imageSource string) bool
	FetchImage(imageSource string, timeout time.Duration) (*Image, error)
}

var imageBackends []ImageBackend

func RegisterImageBackend(fetcher ImageBackend) {
	imageBackends = append(imageBackends, fetcher)
}

func FetchImage(imageSource string, timeout time.Duration) (*Image, error) {
	for _, fetcher := range imageBackends {
		if fetcher.CanHandle(imageSource) {
			return fetcher.FetchImage(imageSource, timeout)
		}
	}

	config := ReadConfig()
	if config.ImageBackend != "http" {
		for _, fetcher := range imageBackends {
			if fetcher.Name() == config.ImageBackend {
				return fetcher.FetchImage(imageSource, timeout)
			}
		}
	}

	return nil, NewStatusError(400, "Unsupported image source: "+imageSource)
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
