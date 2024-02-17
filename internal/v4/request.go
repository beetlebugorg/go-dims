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
	"io"
	"log/slog"
	"net/http"
	"strings"
)

var commandsV4 = map[string]dims.Operation{
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
	Timestamp int32
}

func NewRequest(r *http.Request, config dims.Config) *RequestV4 {
	var timestamp int32
	fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)

	h := md5.New()
	io.WriteString(h, r.PathValue("clientId"))
	io.WriteString(h, r.PathValue("commands"))
	io.WriteString(h, r.URL.Query().Get("url"))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	var commands []dims.Command
	parsedCommands := strings.Split(r.PathValue("commands"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, dims.Command{
			Name:      command,
			Args:      args,
			Operation: commandsV4[command],
		})
	}

	return &RequestV4{
		dims.Request{
			Id:        requestHash,
			Config:    config,
			ClientId:  r.PathValue("clientId"),
			ImageUrl:  r.URL.Query().Get("url"),
			Commands:  commands,
			Signature: r.PathValue("signature"),
		},
		timestamp,
	}
}

// ValidateSignature verifies the signature of the image resize is valid.
func (r *RequestV4) ValidateSignature() bool {
	slog.Debug("verifySignature", "url", r.ImageUrl)

	algorithm := NewMD5(r.Config.Signing.SigningKey, r.Timestamp)
	signature := algorithm.Sign(r.Commands, r.ImageUrl)

	if bytes.Equal([]byte(signature), []byte(r.Signature)) {
		return true
	}

	slog.Error("verifySignature failed.", "expected", signature, "got", r.Signature)

	return false
}
