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

package dims

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
	Image      string `env:"DIMS_ERROR_IMAGE" envDefault:""`
	Background string `env:"DIMS_ERROR_BACKGROUND" envDefault:"#5ADAFD"`
}

type Signing struct {
	SigningAlgorithm string `env:"DIMS_SIGNING_ALGORITHM" envDefault:"hmac-sha256"`
	SigningKey       string `env:"DIMS_SIGNING_KEY,notEmpty"`
}

type OutputFormat struct {
	OutputFormat string   `env:"DIMS_OUTPUT_FORMAT"`
	Exclude      []string `env:"DIMS_OUTPUT_FORMAT_EXCLUDE"`
}

type Timeout struct {
	Download int `env:"DIMS_DOWNLOAD_TIMEOUT" envDefault:"3000"`
}

type Options struct {
	StripMetadata      bool `env:"DIMS_STRIP_METADATA" envDefault:"true"`
	IncludeDisposition bool `env:"DIMS_INCLUDE_DISPOSITION" envDefault:"false"`
}

type ArgumentConfig struct {
	DevelopmentMode bool
	DebugMode       bool
}

type EnvironmentConfig struct {
	Timeout
	EdgeControl
	Signing
	Error
	OriginCacheControl
	OutputFormat
	Options
}

type Config struct {
	EnvironmentConfig
	ArgumentConfig
}

func ReadConfig() EnvironmentConfig {
	cfg := EnvironmentConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return cfg
}
