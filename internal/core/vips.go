package core

import (
	"log/slog"

	"github.com/davidbyttow/govips/v2/vips"
)

// StartVips starts libvips with the configured cache and concurrency limits.
// Call it after the logger is configured, so the effective values reach the
// log in the configured format.
func StartVips(config *Config) {
	vipsConfig := &vips.Config{
		ConcurrencyLevel: config.ConcurrencyLevel,
		MaxCacheFiles:    config.MaxCacheFiles,
		MaxCacheMem:      config.MaxCacheMem,
		MaxCacheSize:     config.MaxCacheSize,
	}

	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(vipsConfig)

	slog.Debug("libvips started",
		"version", vips.Version,
		"concurrency", vipsConfig.ConcurrencyLevel,
		"maxCacheMem", vipsConfig.MaxCacheMem,
		"maxCacheSize", vipsConfig.MaxCacheSize,
		"maxCacheFiles", vipsConfig.MaxCacheFiles)
}
