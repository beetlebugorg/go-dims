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
	"github.com/alecthomas/kong"
	"github.com/beetlebugorg/go-dims/pkg/dims"
	"log/slog"
	"net/http"
	"os"
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
	Dev      bool   `help:"Enable development mode." default:"false"`
	ImageURL string `arg:"" name:"imageUrl" help:"Image URL to sign. For v4 urls place any value in the signature position in the URL."`
}

func (s *SignCmd) Run() error {
	signedUrl, err := dims.SignUrl(s.ImageURL, s.Dev)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", signedUrl)

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
