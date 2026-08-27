package core

import (
	"errors"
	"fmt"

	"github.com/caarlos0/env/v10"
)

type OriginCacheControl struct {
	UseOrigin bool `env:"DIMS_CACHE_CONTROL_USE_ORIGIN" envDefault:"false"`
	Min       int  `env:"DIMS_CACHE_CONTROL_MIN" envDefault:"0"`
	Max       int  `env:"DIMS_CACHE_CONTROL_MAX" envDefault:"0"`
	Default   int  `env:"DIMS_CACHE_CONTROL_DEFAULT" envDefault:"31536000"`
	Error     int  `env:"DIMS_CACHE_CONTROL_ERROR" envDefault:"60"`
}

type EdgeControl struct {
	DownstreamTtl int `env:"DIMS_EDGE_CONTROL_DOWNSTREAM_TTL" envDefault:"0"`
}

type Error struct {
	Background string `env:"DIMS_ERROR_BACKGROUND" envDefault:"#5ADAFD"`
}

type Signing struct {
	SigningKey string `env:"DIMS_SIGNING_KEY"`

	// Compat set to "legacy" also accepts signatures that cover only the
	// parameters named by _keys. Use it to migrate existing URLs, then clear
	// it. Empty means strict.
	Compat string `env:"DIMS_SIGNING_COMPAT"`
}

type OutputFormat struct {
	Default  string   `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`
	Excluded []string `env:"DIMS_EXCLUDED_OUTPUT_FORMATS"`
}

type Server struct {
	// All values are milliseconds. Zero disables the timeout, which is the
	// standard library default and the reason a slow client could previously
	// hold a connection open for as long as it liked.
	ReadHeaderTimeout int `env:"DIMS_READ_HEADER_TIMEOUT" envDefault:"5000"`
	ReadTimeout       int `env:"DIMS_READ_TIMEOUT" envDefault:"15000"`
	WriteTimeout      int `env:"DIMS_WRITE_TIMEOUT" envDefault:"60000"`
	IdleTimeout       int `env:"DIMS_IDLE_TIMEOUT" envDefault:"120000"`
	ShutdownTimeout   int `env:"DIMS_SHUTDOWN_TIMEOUT" envDefault:"30000"`
	MaxHeaderBytes    int `env:"DIMS_MAX_HEADER_BYTES" envDefault:"65536"`
}

type Limits struct {
	// Cost is linear in output pixels, measured at roughly 30 megapixels per
	// second per thread. These caps bound the work one request can ask for.
	// Zero disables a check.
	MaxSourcePixels int `env:"DIMS_MAX_SOURCE_PIXELS" envDefault:"100000000"`
	MaxOutputPixels int `env:"DIMS_MAX_OUTPUT_PIXELS" envDefault:"50000000"`

	// MaxConcurrent caps how many images are processed at once. Zero derives
	// a value from the CPU count. A negative value removes the limit.
	MaxConcurrent int `env:"DIMS_MAX_CONCURRENT" envDefault:"0"`

	// MaxConcurrentWait is how long a request queues for a slot before it is
	// refused, in milliseconds.
	MaxConcurrentWait int `env:"DIMS_MAX_CONCURRENT_WAIT" envDefault:"5000"`

	// MaxSourceBytes caps the size of one source image, in bytes. The pixel
	// caps do not bound memory, because the whole body is resident before any
	// pixel check runs. Zero disables the check.
	MaxSourceBytes int `env:"DIMS_MAX_SOURCE_BYTES" envDefault:"67108864"`

	// MaxDownloadConcurrent caps how many source downloads run at once. Each
	// download holds a whole source body, so this and MaxSourceBytes together
	// bound the memory a burst can reach. Zero derives a value from the CPU
	// count. A negative value removes the limit.
	MaxDownloadConcurrent int `env:"DIMS_MAX_DOWNLOAD_CONCURRENT" envDefault:"0"`
}

type Network struct {
	// AllowedHosts limits which hosts an image may be fetched from. An entry
	// starting with a dot matches any subdomain. An empty list allows any host
	// that passes the address check.
	AllowedHosts []string `env:"DIMS_ALLOWED_HOSTS"`

	// AllowPrivateNetworks permits connections to loopback, link local,
	// private, and other non-public addresses. Keep it false unless the origin
	// sits inside the same network.
	AllowPrivateNetworks bool `env:"DIMS_ALLOW_PRIVATE_NETWORKS" envDefault:"false"`

	// MaxRedirects caps how many redirects one fetch will follow.
	MaxRedirects int `env:"DIMS_MAX_REDIRECTS" envDefault:"3"`
}

type Timeout struct {
	Download int `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"3000"`
}

type Options struct {
	StripMetadata      bool `env:"DIMS_STRIP_METADATA" envDefault:"true"`
	IncludeDisposition bool `env:"DIMS_INCLUDE_DISPOSITION" envDefault:"false"`
}

type JpegCompression struct {
	Quality            int  `env:"DIMS_JPEG_QUALITY" envDefault:"80"`
	Interlace          bool `env:"DIMS_JPEG_INTERLACE" envDefault:"false"`
	OptimizeCoding     bool `env:"DIMS_JPEG_OPTIMIZE_CODING" envDefault:"true"`
	SubsampleMode      bool `env:"DIMS_JPEG_SUBSAMPLE_MODE" envDefault:"true"`
	TrellisQuant       bool `env:"DIMS_JPEG_TRELLIS_QUANT" envDefault:"false"`
	OvershootDeringing bool `env:"DIMS_JPEG_OVERSHOOT_DERINGING" envDefault:"false"`
	OptimizeScans      bool `env:"DIMS_JPEG_OPTIMIZE_SCANS" envDefault:"false"`
	QuantTable         int  `env:"DIMS_JPEG_QUANT_TABLE" envDefault:"3"`
}

type PngCompression struct {
	Quality     int  `env:"DIMS_PNG_QUALITY" envDefault:"80"`
	Interlace   bool `env:"DIMS_PNG_INTERLACE" envDefault:"false"`
	Compression int  `env:"DIMS_PNG_COMPRESSION" envDefault:"4"`
}

type WebpCompression struct {
	Quality         int    `env:"DIMS_WEBP_QUALITY" envDefault:"80"`
	Compression     string `env:"DIMS_WEBP_COMPRESSION" envDefault:"lossy"`
	ReductionEffort int    `env:"DIMS_WEBP_REDUCTION_EFFORT" envDefault:"4"`
}

type ImageOutputOptions struct {
	Jpeg JpegCompression
	Png  PngCompression
	Webp WebpCompression
}

type Source struct {
	Default string   `env:"DIMS_DEFAULT_SOURCE_BACKEND" envDefault:"http"`
	Allowed []string `env:"DIMS_ALLOWED_SOURCE_BACKENDS" envDefault:"http"`
}

type S3 struct {
	Region string `env:"DIMS_S3_REGION" envDefault:""`
	Bucket string `env:"DIMS_S3_BUCKET" envDefault:""`
	Prefix string `env:"DIMS_S3_PREFIX" envDefault:""`
}

type FileSource struct {
	BaseDir string `env:"DIMS_FILE_BASE_DIR" envDefault:"./resources"`
}

type Config struct {
	BindAddress     string `env:"DIMS_BIND_ADDRESS" envDefault:":8080"`
	DevelopmentMode bool   `env:"DIMS_DEVELOPMENT_MODE" envDefault:"false"`
	DebugMode       bool   `env:"DIMS_DEBUG_MODE" envDefault:"false"`
	LogFormat       string `env:"DIMS_LOG_FORMAT" envDefault:"text"`
	EtagAlgorithm   string

	Timeout
	Limits
	Server
	Network
	EdgeControl
	Signing
	Error
	OriginCacheControl
	OutputFormat
	Options
	ImageOutputOptions
}

var config *Config

// startupErrors collects configuration failures raised while packages
// initialise. They cannot be logged where they happen, because logging is not
// configured until the command runs, and printing them bypasses the
// configured log format. Validate reports them instead.
var startupErrors []error

// RecordStartupError stores a configuration failure for Validate to report.
// Only safe to call during package initialisation, which is single threaded.
func RecordStartupError(err error) {
	if err != nil {
		startupErrors = append(startupErrors, err)
	}
}

func init() {
	config = &Config{}
	RecordStartupError(env.Parse(config))
}

func ReadConfig() *Config {
	return config
}

// Validate reports settings that would otherwise fail quietly at request time.
func (c *Config) Validate() error {
	if len(startupErrors) > 0 {
		return fmt.Errorf("invalid configuration: %w", errors.Join(startupErrors...))
	}

	if c.OutputFormat.Default == "" {
		return nil
	}

	if _, ok := ImageTypes[c.OutputFormat.Default]; !ok {
		return fmt.Errorf("DIMS_DEFAULT_OUTPUT_FORMAT %q is not a supported format", c.OutputFormat.Default)
	}

	return nil
}
