package dims

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	DownloadTimeout          int     `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"3000"`
	ImagemagickTimeout       int     `env:"DIMS_IMAGEMAGICK_TIMEOUT" envDefault:"3000"`
	PlaceholderImageUrl      string  `env:"DIMS_PLACEHOLDER_IMAGE_URL"`
	PlaceholderImageExpire   int     `env:"DIMS_PLACEHOLDER_IMAGE_EXPIRE" envDefault:"60"`
	DefaultExpire            int     `env:"DIMS_DEFAULT_EXPIRE" envDefault:"31536000"`
	StripMetadata            bool    `env:"DIMS_STRIP_METADATA" envDefault:"true"`
	SampleFactor             float32 `env:"DIMS_SAMPLE_FACTOR"`
	IncludeDisposition       bool    `env:"DIMS_INCLUDE_DISPOSITION" envDefault:"false"`
	DisableEncodedFetch      bool    `env:"DIMS_DISABLE_ENCODED_FETCH" envDefault:"false"`
	DefaultOutputFormat      string  `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`
	SecretKey                string  `env:"DIMS_SECRET_KEY"`
	DefaultImagePrefix       string  `env:"DIMS_DEFAULT_IMAGE_PREFIX"`
	CacheControlMaxAge       int     `env:"DIMS_CACHE_CONTROL_MAX_AGE" envDefault:"86400"`
	EdgeControlDownstreamTtl int     `env:"DIMS_EDGE_CONTROL_DOWNSTREAM_TTL" envDefault:"-1"`
	TrustSrc                 bool    `env:"DIMS_TRUST_SRC" envDefault:"false"`
	MinSrcCacheControl       int     `env:"DIMS_MIN_SRC_CACHE_CONTROL" envDefault:"-1"`
	MaxSrcCacheControl       int     `env:"DIMS_MAX_SRC_CACHE_CONTROL" envDefault:"-1"`
	//MagickSizeType area_size
	//MagickSizeType memory_size
	//MagickSizeType map_size
	//MagickSizeType disk_size

}

func ReadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return cfg
}
