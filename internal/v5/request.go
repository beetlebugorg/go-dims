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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/v4"
	"log/slog"
	"net/http"
	"strings"
)

var CommandsV5 = map[string]dims.Operation{
	"crop":       v4.CropCommand,
	"resize":     v4.ResizeCommand,
	"strip":      v4.StripMetadataCommand,
	"format":     v4.FormatCommand,
	"quality":    v4.QualityCommand,
	"sharpen":    v4.SharpenCommand,
	"brightness": v4.BrightnessCommand,
	"flipflop":   v4.FlipFlopCommand,
	"sepia":      v4.SepiaCommand,
	"grayscale":  v4.GrayscaleCommand,
	"autolevel":  v4.AutolevelCommand,
	"invert":     v4.InvertCommand,
	"rotate":     v4.RotateCommand,
	"thumbnail":  v4.ThumbnailCommand,
	"gravity":    GravityCommand,
}

type RequestV5 struct {
	dims.Request
}

func NewRequest(r *http.Request, config dims.Config) *RequestV5 {
	h := sha256.New()
	h.Write([]byte(r.PathValue("clientId")))
	h.Write([]byte(r.PathValue("commands")))
	h.Write([]byte(r.URL.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestV5{
		dims.Request{
			Id:        requestHash,
			Config:    config,
			ClientId:  r.URL.Query().Get("clientId"),
			ImageUrl:  r.URL.Query().Get("url"),
			Commands:  dims.ParseCommands(r.PathValue("commands"), CommandsV5),
			Signature: r.URL.Query().Get("sig"),
		},
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
	mac := hmac.New(sha256.New, []byte(r.Config.SigningKey))
	for _, command := range r.Commands {
		sanitizedArgs := strings.ReplaceAll(command.Args, " ", "+")

		mac.Write([]byte(command.Name))
		mac.Write([]byte("/"))
		mac.Write([]byte(sanitizedArgs))
	}
	mac.Write([]byte(r.ImageUrl))

	return mac.Sum(nil)[0:31]
}
