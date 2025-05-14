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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/beetlebugorg/go-dims/internal/dims/core"
	"github.com/beetlebugorg/go-dims/pkg/dims"
)

// - dims serve

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

type EncryptionCmd struct {
	URL string `arg:"" help:"The URL to encrypt."`
}

func (e *EncryptionCmd) Run() error {
	if e.URL == "" {
		return errors.New("URL is required")
	}

	config := core.ReadConfig()

	result, err := dims.EncryptURL(config.SigningKey, e.URL)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	fmt.Println("Encrypted URL:", result)

	return nil
}

type DecryptionCmd struct {
	URL string `arg:"" help:"The URL to decrypt."`
}

func (e *DecryptionCmd) Run() error {
	if e.URL == "" {
		return errors.New("URL is required")
	}

	config := core.ReadConfig()

	result, err := dims.DecryptURL(config.SigningKey, e.URL)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	fmt.Println("Decrypted URL:", result)

	return nil
}

type HealthCheckCmd struct {
	Port int `help:"Port to check." default:"8080"`
}

func (h *HealthCheckCmd) Run() error {
	url := fmt.Sprintf("http://localhost:%d/healthz", h.Port)
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %v", err)
	}

	return nil
}

var CLI struct {
	Serve       ServeCmd       `cmd:"" help:"Runs the DIMS service."`
	Encrypt     EncryptionCmd  `cmd:"" help:"Encrypt an eurl."`
	Decrypt     DecryptionCmd  `cmd:"" help:"Decrypt an eurl."`
	HealthCheck HealthCheckCmd `cmd:"" help:"Check the health of the DIMS service."`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	if err != nil {
		os.Exit(1)
	}
}
