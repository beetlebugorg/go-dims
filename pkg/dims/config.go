package dims

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	DownloadTimeout          int     `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"60000"`
	ImagemagickTimeout       int     `env:"DIMS_IMAGEMAGICK_TIMEOUT" envDefault:"20000"`
	NoImageUrl               string  `env:"DIMS_NO_IMAGE_URL"`
	NoImageExpire            int     `env:"DIMS_NO_IMAGE_EXPIRE"`
	DefaultExpire            int64   `env:"DIMS_DEFAULT_EXPIRE envDefault:"31536000"`
	StripMetadata            bool    `env:"DIMS_STRIP_METADATA envDefault:"true"`
	SampleFactor             float32 `env:"DIMS_SAMPLE_FACTOR"`
	IncludeDisposition       bool    `env:"DIMS_INCLUDE_DISPOSITION" envDefault:"false"`
	DisableEncodedFetch      bool    `env:"DIMS_DISABLE_ENCODED_FETCH"`
	DefaultOutputFormat      string  `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`
	SecretKey                string  `env:"DIMS_SECRET_KEY"`
	MaxExpiryPeriod          int     `env:"DIMS_MAX_EXPIRY_PERIOD"`
	DefaultImagePrefix       string  `env:"DIMS_DEFAULT_IMAGE_PREFIX"`
	CacheControlMaxAge       int     `env:"DIMS_CACHE_CONTROL_MAX_AGE"`
	EdgeControlDownstreamTtl int     `env:"DIMS_EDGE_CONTROL_DOWNSTREAM_TTL"`
	TrustSrc                 bool    `env:"DIMS_TRUST_SRC"`
	MinSrcCacheControl       int     `env:"DIMS_MIN_SRC_CACHE_CONTROL"`
	MaxSrcCacheControl       int     `env:"DIMS_MAX_SRC_CACHE_CONTROL"`
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
