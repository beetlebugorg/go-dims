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

package main

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/pkg/signing"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kong"
	"github.com/beetlebugorg/go-dims/pkg/dims"
)

type ServeCmd struct {
	Bind  string `help:"Bind address to serve on." default:":8080"`
	Debug bool   `help:"Enable debug mode." default:"false"`
	Dev   bool   `help:"Enable development mode." default:"false"`
}

func (s *ServeCmd) Run() error {
	var opts *slog.HandlerOptions
	if s.Debug {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	err := http.ListenAndServe(s.Bind, dims.NewHandler(s.Debug, s.Dev))
	if err != nil {
		slog.Error("Server failed.", "error", err)
		return err
	}
	return nil
}

type SignCmd struct {
	SigningAlgorithm string `help:"SigningAlgorithm to use." default:"hmac-sha256"`
	Timestamp        int32  `arg:"" name:"timestamp" help:"Expiration timestamp" type:"timestamp"`
	Secret           string `arg:"" name:"secret" help:"Secret key."`
	Commands         string `arg:"" name:"commands" help:"Commands to sign."`
	ImageURL         string `arg:"" name:"image_url" help:"Image URL to sign."`
}

func (s *SignCmd) Run() error {
	var algorithm signing.SignatureAlgorithm
	if s.SigningAlgorithm == "md5" {
		algorithm = signing.NewMD5(s.Secret, s.Timestamp)
	} else if s.SigningAlgorithm == "hmac-sha256" {
		algorithm = signing.NewHmacSha256(s.Secret)
	}

	fmt.Printf("%s\n", algorithm.Sign(s.Commands, s.ImageURL))

	return nil
}

var CLI struct {
	Serve ServeCmd `cmd:"" help:"Runs the DIMS service."`
	Sign  SignCmd  `cmd:"" help:"Signs the given image URL."`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	if err != nil {
		return
	}
}
