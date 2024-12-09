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

package lambda

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
)

var CommandsLambda = map[string]dims.VipsOperation{
	"crop":       dims.CropCommand,
	"resize":     dims.ResizeCommand,
	"strip":      dims.StripMetadataCommand,
	"format":     dims.FormatCommand,
	"quality":    dims.QualityCommand,
	"sharpen":    dims.SharpenCommand,
	"brightness": dims.BrightnessCommand,
	"flipflop":   dims.FlipFlopCommand,
	"sepia":      dims.SepiaCommand,
	"grayscale":  dims.GrayscaleCommand,
	"autolevel":  dims.AutolevelCommand,
	"invert":     dims.InvertCommand,
	"rotate":     dims.RotateCommand,
	"thumbnail":  dims.ThumbnailCommand,
}

type RequestLambdaFunctionURL struct {
	v5.RequestV5
}

func NewLambdaFunctionURLRequest(event events.LambdaFunctionURLRequest, config dims.Config) *RequestLambdaFunctionURL {
	u, err := url.Parse(event.RawPath + "?" + event.RawQueryString)
	if err != nil {
		return nil
	}

	slog.Info("NewLambdaFunctionURLRequest", "event", event)

	// /v5/dims/{commands...}
	rawCommands := strings.TrimLeft(event.RawPath, "/v5/dims/")

	h := sha256.New()
	h.Write([]byte(u.Query().Get("clientId")))
	h.Write([]byte(rawCommands))
	h.Write([]byte(u.Query().Get("url")))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return &RequestLambdaFunctionURL{
		RequestV5: v5.RequestV5{
			Request: dims.Request{
				Id:          requestHash,
				Config:      config,
				ClientId:    u.Query().Get("clientId"),
				ImageUrl:    u.Query().Get("url"),
				RawCommands: rawCommands,
				Signature:   u.Query().Get("sig"),
			},
		},
	}
}
