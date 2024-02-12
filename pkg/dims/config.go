package dims

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	DownloadTimeout    int `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"3000"`
	ImagemagickTimeout int `env:"DIMS_IMAGEMAGICK_TIMEOUT" envDefault:"3000"`

	//apr_hash_t  *clients
	//apr_table_t *ignore_default_output_format

	NoImageUrl          string  `env:"DIMS_NO_IMAGE_URL"`
	NoImageExpire       int     `env:"DIMS_NO_IMAGE_EXPIRE"`
	DefaultExpire       int64   `env:"DIMS_DEFAULT_EXPIRE"`
	StripMetadata       bool    `env:"DIMS_STRIP_METADATA"`
	SampleFactor        float64 `env:"DIMS_SAMPLE_FACTOR"`
	IncludeDisposition  bool    `env:"DIMS_INCLUDE_DISPOSITION"`
	DisableEncodedFetch bool    `env:"DIMS_DISABLE_ENCODED_FETCH"`
	DefaultOutputFormat string  `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`

	//MagickSizeType area_size
	//MagickSizeType memory_size
	//MagickSizeType map_size
	//MagickSizeType disk_size

	SecretKey          string `env:"DIMS_SECRET_KEY"`
	MaxExpiryPeriod    int64  `env:"DIMS_MAX_EXPIRY_PERIOD"`
	DefaultImagePrefix string `env:"DIMS_DEFAULT_IMAGE_PREFIX"`

	CacheControlMaxAge       int64 `env:"DIMS_CACHE_CONTROL_MAX_AGE"`
	EdgeControlDownstreamTtl int64 `env:"DIMS_EDGE_CONTROL_DOWNSTREAM_TTL"`
	TrustSrc                 int64 `env:"DIMS_TRUST_SRC"`
	MinSrcCacheControl       int64 `env:"DIMS_MIN_SRC_CACHE_CONTROL"`
	MaxSrcCacheControl       int64 `env:"DIMS_MAX_SRC_CACHE_CONTROL"`
}

func ReadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return cfg
}
