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

package v4

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
	"net/http"
	"strings"
)

var commandsV4 = map[string]MagickOperation{
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

type RequestV4 struct {
	dims.Request
	Timestamp     int32
	TrailingSlash bool
	mw            *imagick.MagickWand
}

func init() {
	imagick.Initialize()
}

func NewRequest(r *http.Request, config dims.Config) *RequestV4 {
	var timestamp int32
	n, err := fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)
	if err != nil || n != 1 {
		timestamp = 0
	}

	h := md5.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestV4{
		dims.Request{
			Id:          requestHash,
			Config:      config,
			ClientId:    r.PathValue("clientId"),
			ImageUrl:    r.URL.Query().Get("url"),
			RawCommands: r.PathValue("commands"),
			Signature:   r.PathValue("signature"),
		},
		timestamp,
		strings.HasSuffix(r.URL.Path, "/"),
		nil,
	}
}

// Parse through the requested commands and set the optimal image size on the MagicWand.
//
// This is used while reading an image to improve
// performance when generating thumbnails from very
// large images.
//
// An example speed is taking 1817x3000 sized image and
// reducing it to a 78x110 thumbnail:
//
//	without MagickSetSize: 396ms
//	with MagickSetSize:    105ms
func (r *RequestV4) setOptimalImageSize() {
	for _, command := range r.Commands() {
		if command.Name == "thumbnail" || command.Name == "resize" {
			var rect imagick.RectangleInfo
			flags := imagick.ParseAbsoluteGeometry(command.Args, &rect)

			if (flags&imagick.WIDTHVALUE != 0) &&
				(flags&imagick.HEIGHTVALUE != 0) &&
				(flags&imagick.PERCENTVALUE == 0) {

				if err := r.mw.SetSize(rect.Width, rect.Height); err != nil {
					slog.Error("setOptimalImageSize failed.", "error", err)
				}

				return
			}
		}
	}
}

// ValidateSignature verifies the signature of the image resize is valid.
func (r *RequestV4) ValidateSignature() bool {
	slog.Debug("verifySignature", "url", r.ImageUrl)

	signature := r.Sign()

	if bytes.Equal([]byte(signature), []byte(r.Signature)) {
		return true
	}

	slog.Error("verifySignature failed.", "expected", signature, "got", r.Signature)

	return false
}

// ProcessImage will execute the commands on the image.
func (r *RequestV4) ProcessImage() (string, []byte, error) {
	slog.Debug("executeImagemagick")

	r.mw = imagick.NewMagickWand()
	mw := r.mw

	// Read the image.
	r.setOptimalImageSize()
	err := mw.ReadImageBlob(r.SourceImage.Bytes)
	if err != nil {
		return "", nil, err
	}

	// Convert image to RGB from CMYK.
	if mw.GetImageColorspace() == imagick.COLORSPACE_CMYK {
		profiles := mw.GetImageProfiles("icc")
		if profiles != nil {
			if err := mw.ProfileImage("ICC", CmykIccProfile); err != nil {
				return "", nil, err
			}
		}

		if err := mw.ProfileImage("ICC", RgbIccProfile); err != nil {
			return "", nil, err
		}

		err = mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
		if err != nil {
			return "", nil, err
		}
	}

	// Flip image orientation, if needed.
	if err := mw.AutoOrientImage(); err != nil {
		return "", nil, err
	}

	// Execute the commands.
	stripMetadata := true
	formatProvided := false

	for _, command := range r.Commands() {
		if command.Name == "strip" && command.Args == "true" {
			stripMetadata = false
		}

		if command.Name == "format" {
			formatProvided = true
		}

		if err := r.ProcessCommand(command); err != nil {
			return "", nil, err
		}

		mw.MergeImageLayers(imagick.IMAGE_LAYER_TRIM_BOUNDS)
	}

	// Strip metadata. (if not already stripped)
	if stripMetadata && r.Config.StripMetadata {
		if err := mw.StripImage(); err != nil {
			return "", nil, err
		}
	}

	// Set output format if not provided in the request.
	if !formatProvided && r.Config.OutputFormat.OutputFormat != "" {
		format := strings.ToLower(mw.GetImageFormat())
		if !contains(r.Config.OutputFormat.Exclude, format) {
			if err := mw.SetImageFormat(r.Config.OutputFormat.OutputFormat); err != nil {
				return "", nil, err
			}
		}
	}

	mw.ResetIterator()

	return mw.GetImageFormat(), mw.GetImagesBlob(), nil
}

// Sign returns a signed string using the MD5 algorithm.
func (r *RequestV4) Sign() string {
	sanitizedCommands := strings.ReplaceAll(r.Request.RawCommands, " ", "+")
	sanitizedCommands = strings.Trim(sanitizedCommands, "/")

	// This makes the signing algorithm compatible with mod-dims.
	if r.TrailingSlash {
		sanitizedCommands += "/"
	}

	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d", r.Timestamp)))
	hash.Write([]byte(r.Config.Signing.SigningKey))
	hash.Write([]byte(sanitizedCommands))
	hash.Write([]byte(r.Request.ImageUrl))

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}

func (r *RequestV4) ProcessCommand(command dims.Command) error {
	if operation, ok := commandsV4[command.Name]; ok {
		return operation(r.mw, command.Args)
	}

	return fmt.Errorf("command not found: %s", command.Name)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
