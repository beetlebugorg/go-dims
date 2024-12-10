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
	"net/http"

	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/davidbyttow/govips/v2/vips"
)

func NewHandler(debug bool, dev bool) http.Handler {
	environmentConfig := dims.ReadConfig()

	mux := http.NewServeMux()

	vips.LoggingSettings(nil, vips.LogLevelError)

	vips.Startup(nil)

	// v4 endpoint
	v4Arguments := "{clientId}/{signature}/{timestamp}/{commands...}"
	v4Handler := func(w http.ResponseWriter, r *http.Request) {
		config := dims.Config{
			EnvironmentConfig: environmentConfig,
			DevelopmentMode:   dev,
			DebugMode:         debug,
			EtagAlgorithm:     "md5",
		}

		if debug {
			fmt.Printf("config: %+v\n", config)
		}

		request, err := dims.ParseAndValidV4Request(r, config)
		if err != nil {
			// send error
			return
		}

		dims.Handler(*request, config, w)
	}
	mux.HandleFunc(fmt.Sprintf("/v4/dims/%s", v4Arguments), v4Handler)

	// v5 endpoint
	mux.HandleFunc("/v5/dims/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			config := dims.Config{
				EnvironmentConfig: environmentConfig,
				DevelopmentMode:   dev,
				DebugMode:         debug,
				EtagAlgorithm:     "hmac-sha256",
			}

			if debug {
				fmt.Printf("config: %+v\n", config)
			}

			request, err := dims.ParseAndValidV5Request(r, config)
			if err != nil {
				// send error
				return
			}

			dims.Handler(*request, config, w)
		})

	mux.HandleFunc("/dims-status",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(environmentConfig, debug, dev, w, r)
		})

	return mux
}
