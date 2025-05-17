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
	"github.com/beetlebugorg/go-dims/internal/gox/imagex/colorx"
	"github.com/caarlos0/env/v10"
	"golang.org/x/exp/slices"
	"log/slog"
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

type SourceBackend interface {
	Name() string
	CanHandle(imageSource string) bool
	FetchImage(imageSource string, timeout time.Duration) (*Image, error)
}

var sourceBackends []SourceBackend
var allSourceBackends []SourceBackend
var defaultSourceBackend SourceBackend

func RegisterImageBackend(sourceBackend SourceBackend) {
	config := Source{}
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}

	allSourceBackends = append(allSourceBackends, sourceBackend)

	allowed := config.Allowed

	if slices.Contains(allowed, sourceBackend.Name()) {
		slog.Info("Registering image backend", "name", sourceBackend.Name())
		sourceBackends = append(sourceBackends, sourceBackend)
	} else if config.Default == sourceBackend.Name() {
		slog.Info("Registering image backend", "name", sourceBackend.Name())
		sourceBackends = append(sourceBackends, sourceBackend)
	} else {
		slog.Warn("Image backend not registered", "name", sourceBackend.Name(), "reason", "not in allowed list")
	}

	if sourceBackend.Name() == config.Default {
		defaultSourceBackend = sourceBackend
	}
}

func ErrorImage(color string) (*vips.ImageRef, error) {
	errorImage, err := vips.Black(512, 512)
	if err != nil {
		return nil, err
	}

	if err := errorImage.BandJoinConst([]float64{0, 0, 255}); err != nil {
		return nil, err
	}

	backgroundColor, err := colorx.ParseHexColor(color)
	if err != nil {
		return nil, err
	}

	red, green, blue, _ := backgroundColor.RGBA()
	redI := float64(red) / 65535 * 255
	greenI := float64(green) / 65535 * 255
	blueI := float64(blue) / 65535 * 255

	if err := errorImage.Linear([]float64{0, 0, 0, 0}, []float64{redI, greenI, blueI, 255}); err != nil {
		return nil, err
	}

	if err := errorImage.Cast(vips.BandFormatUchar); err != nil {
		return nil, err
	}

	return errorImage, nil
}

func FetchImage(imageSource string, timeout time.Duration) (*Image, error) {
	// Check which source backend can handle the image source first, if
	// none are found, use the default source backend.
	for _, sourceBackend := range sourceBackends {
		if sourceBackend.CanHandle(imageSource) {
			return sourceBackend.FetchImage(imageSource, timeout)
		}
	}

	// If a default is set and another backend could have handled it but isn't
	// allowed, don't use the default backend.
	if defaultSourceBackend != nil && defaultSourceBackend.Name() != "http" || defaultSourceBackend == nil {
		for _, sourceBackend := range allSourceBackends {
			if sourceBackend.CanHandle(imageSource) && sourceBackend.Name() != defaultSourceBackend.Name() {
				return nil, NewStatusError(400, "Supported image source '"+sourceBackend.Name()+
					"' is not allowed: "+
					imageSource)
			}
		}

		return defaultSourceBackend.FetchImage(imageSource, timeout)
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
