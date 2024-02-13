package dims

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	DownloadTimeout            int      `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"3000"`
	ImagemagickTimeout         int      `env:"DIMS_IMAGEMAGICK_TIMEOUT" envDefault:"3000"`
	PlaceholderBackground      string   `env:"DIMS_PLACEHOLDER_BACKGROUND" envDefault:"#5ADAFD"`
	PlaceholderImageExpire     int      `env:"DIMS_PLACEHOLDER_IMAGE_EXPIRE" envDefault:"60"`
	DefaultExpire              int      `env:"DIMS_DEFAULT_EXPIRE" envDefault:"31536000"`
	StripMetadata              bool     `env:"DIMS_STRIP_METADATA" envDefault:"true"`
	IncludeDisposition         bool     `env:"DIMS_INCLUDE_DISPOSITION" envDefault:"false"`
	DefaultOutputFormat        string   `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`
	IgnoreDefaultOutputFormats []string `env:"DIMS_IGNORE_DEFAULT_OUTPUT_FORMATS"`
	SecretKey                  string   `env:"DIMS_SECRET_KEY"`
	DefaultImagePrefix         string   `env:"DIMS_DEFAULT_IMAGE_PREFIX"`
	CacheControlMaxAge         int      `env:"DIMS_CACHE_CONTROL_MAX_AGE" envDefault:"86400"`
	EdgeControlDownstreamTtl   int      `env:"DIMS_EDGE_CONTROL_DOWNSTREAM_TTL" envDefault:"-1"`
	TrustSrc                   bool     `env:"DIMS_TRUST_SRC" envDefault:"false"`
	MinSrcCacheControl         int      `env:"DIMS_MIN_SRC_CACHE_CONTROL" envDefault:"-1"`
	MaxSrcCacheControl         int      `env:"DIMS_MAX_SRC_CACHE_CONTROL" envDefault:"-1"`
}

func ReadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return cfg
}
