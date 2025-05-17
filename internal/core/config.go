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

package core

import (
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
	SigningKey string `env:"DIMS_SIGNING_KEY,notEmpty"`
}

type OutputFormat struct {
	Default  string   `env:"DIMS_DEFAULT_OUTPUT_FORMAT"`
	Excluded []string `env:"DIMS_EXCLUDED_OUTPUT_FORMATS"`
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
	EtagAlgorithm   string

	Timeout
	EdgeControl
	Signing
	Error
	OriginCacheControl
	OutputFormat
	Options
	ImageOutputOptions
}

func ReadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return cfg
}
