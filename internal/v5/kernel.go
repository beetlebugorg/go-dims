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

package v5

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/http"
	"runtime/trace"
	"strings"
)

var CommandsV5 = map[string]VipsOperation{
	"resize":  ResizeCommand,
	"strip":   StripMetadataCommand,
	"format":  PassThroughCommand,
	"quality": PassThroughCommand,
	"crop":    CropCommand,
}

type RequestV5 struct {
	dims.Request
	format    string
	strip     bool
	vipsImage *vips.ImageRef
}

func NewRequest(r *http.Request, config dims.Config) *RequestV5 {
	h := sha256.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestV5{
		Request: dims.Request{
			Id:          requestHash,
			Config:      config,
			ClientId:    r.URL.Query().Get("clientId"),
			ImageUrl:    r.URL.Query().Get("url"),
			RawCommands: r.PathValue("commands"),
			Signature:   r.URL.Query().Get("sig"),
		},
		format: config.OutputFormat.OutputFormat,
		strip:  config.StripMetadata,
	}
}

// ValidateSignature verifies the signature of the image resize is valid.
func (r *RequestV5) ValidateSignature() bool {
	slog.Debug("verifySignature", "url", r.ImageUrl)

	expectedSignature := r.Sign()
	gotSignature, err := hex.DecodeString(r.Signature)
	if err != nil {
		slog.Error("verifySignature failed.", "error", err)
		return false
	}

	if hmac.Equal(expectedSignature, gotSignature) {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", hex.EncodeToString(expectedSignature),
		"got", r.Signature)

	return false
}

// Sign returns a signed string using HMAC-SHA256-128.
func (r *RequestV5) Sign() []byte {
	sanitizedArgs := strings.ReplaceAll(r.Request.RawCommands, " ", "+")

	mac := hmac.New(sha256.New, []byte(r.Config.SigningKey))
	mac.Write([]byte(sanitizedArgs))
	mac.Write([]byte(r.Request.ImageUrl))

	return mac.Sum(nil)[0:31]
}

// ProcessImage will execute the commands on the image.
func (r *RequestV5) ProcessImage() (string, []byte, error) {
	slog.Debug("executeVips")

	importParams := vips.NewImportParams()
	importParams.AutoRotate.Set(true)

	image, err := vips.LoadImageFromBuffer(r.SourceImage.Bytes, importParams)
	if err != nil {
		return "", nil, err
	}

	r.vipsImage = image

	// Execute the commands.
	stripMetadata := r.Config.StripMetadata
	//formatProvided := false

	ctx, task := trace.NewTask(context.Background(), "v5.ProcessImage")
	defer task.End()

	for _, command := range r.Commands() {
		if command.Name == "strip" && command.Args == "true" {
			stripMetadata = false
		}

		//if command.Name == "format" {
		//	formatProvided = true
		//}

		region := trace.StartRegion(ctx, command.Name)
		if err := r.ProcessCommand(command); err != nil {
			return "", nil, err
		}
		region.End()
	}

	if stripMetadata {
		// set strip option in export options
	}

	if image.Format() == vips.ImageTypeJPEG {
		exportParams := vips.NewJpegExportParams()
		exportParams.StripMetadata = true
		exportParams.Quality = 75
		exportParams.OptimizeCoding = true
		mode := vips.VipsForeignSubsampleAuto
		exportParams.SubsampleMode = mode

		imageBytes, _, err := image.ExportJpeg(exportParams)
		if err != nil {
			return "", nil, err
		}

		return vips.ImageTypes[image.Format()], imageBytes, nil
	}

	imageBytes, _, err := image.ExportNative()
	if err != nil {
		return "", nil, err
	}

	return vips.ImageTypes[image.Format()], imageBytes, nil
}

func (r *RequestV5) ProcessCommand(command dims.Command) error {
	if operation, ok := CommandsV5[command.Name]; ok {
		return operation(r.vipsImage, command.Args)
	}

	return fmt.Errorf("command not found: %s", command.Name)
}
