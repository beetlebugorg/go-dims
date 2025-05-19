package main

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/pkg/dims"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/http"
	"os"
)

type ServeCmd struct {
}

func (s *ServeCmd) Run() error {
	config := core.ReadConfig()

	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)

	var opts *slog.HandlerOptions
	if config.DebugMode {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	}

	var logger *slog.Logger
	if config.LogFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	if !config.DevelopmentMode && config.SigningKey == "" {
		slog.Error("Signing key is required in production mode.")
		return fmt.Errorf("signing key is required in production mode")
	}

	err := http.ListenAndServe(config.BindAddress, dims.NewHandler(*config))
	if err != nil {
		slog.Error("Server failed.", "error", err)
		return err
	}
	return nil
}
