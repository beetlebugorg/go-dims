package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/pkg/dims"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	if err := config.Validate(); err != nil {
		slog.Error("Invalid configuration.", "error", err)
		return err
	}

	if !config.DevelopmentMode && config.SigningKey == "" {
		slog.Error("Signing key is required in production mode.")
		return fmt.Errorf("signing key is required in production mode")
	}

	server := newServer(config)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	failed := make(chan error, 1)
	go func() {
		slog.Info("listening", "address", config.BindAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()

	select {
	case err := <-failed:
		slog.Error("Server failed.", "error", err)
		vips.Shutdown()
		return err
	case <-ctx.Done():
		stop()
	}

	// Let requests already in flight finish. Each one holds libvips buffers,
	// so dropping them mid-encode wastes the work and truncates the response.
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), milliseconds(config.ShutdownTimeout))
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Shutdown did not complete.", "error", err)
	}

	vips.Shutdown()

	return err
}

// newServer builds the listener from configuration. Without these timeouts a
// slow client holds a connection, and the libvips buffers attached to its
// request, for as long as it likes.
func newServer(config *core.Config) *http.Server {
	return &http.Server{
		Addr:              config.BindAddress,
		Handler:           dims.NewHandler(*config),
		ReadHeaderTimeout: milliseconds(config.ReadHeaderTimeout),
		ReadTimeout:       milliseconds(config.ReadTimeout),
		WriteTimeout:      milliseconds(config.WriteTimeout),
		IdleTimeout:       milliseconds(config.IdleTimeout),
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}
}

// milliseconds converts a configured value. Zero stays zero, which the http
// package reads as no timeout.
func milliseconds(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}
