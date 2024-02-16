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
	"crypto/sha256"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/v4"
	"io"
	"net/http"
	"strings"
)

var V5_COMMANDS = map[string]dims.Operation{
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
	"gravity":    v4.GravityCommand,
}

func NewRequest(r *http.Request, config dims.Config) *dims.Request {
	var timestamp int32
	fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)

	h := sha256.New()
	io.WriteString(h, r.PathValue("clientId"))
	io.WriteString(h, r.PathValue("commands"))
	io.WriteString(h, r.URL.Query().Get("url"))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	var commands []dims.Command
	parsedCommands := strings.Split(r.PathValue("commands"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, dims.Command{
			Name:      command,
			Args:      args,
			Operation: V5_COMMANDS[command],
		})
	}

	return &dims.Request{
		Id:        hash,
		Config:    config,
		ClientId:  r.PathValue("clientId"),
		ImageUrl:  r.URL.Query().Get("url"),
		Commands:  commands,
		Signature: r.PathValue("signature"),
	}
}
