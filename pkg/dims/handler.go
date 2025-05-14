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
	"log/slog"
	"net/http"

	"github.com/beetlebugorg/go-dims/internal/dims"
	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

func NewHandler(config core.Config) http.Handler {

	mux := http.NewServeMux()

	slog.Debug("startup", "config", config)

	// v4 endpoint
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			config.EtagAlgorithm = "md5"

			request, err := dims.ParseAndValidateV4Request(r, config)
			if err != nil {
				request.SendError(w, err)
				return
			}

			if err := dims.Handler(*request, config, w); err != nil {
				request.SendError(w, err)
				return
			}
		})

	// v5 endpoint
	mux.HandleFunc("/v5/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			config.EtagAlgorithm = "hmac-sha256"

			request, err := dims.ParseAndValidateV5Request(r, config)
			if err != nil {
				request.SendError(w, err)
				return
			}

			if err := dims.Handler(*request, config, w); err != nil {

				request.SendError(w, err)
				return
			}
		})

	mux.HandleFunc("/dims-status/",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, w, r)
		})
	mux.HandleFunc("/healthz",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, w, r)
		})

	return mux
}
