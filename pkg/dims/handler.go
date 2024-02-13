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
	"gopkg.in/gographics/imagick.v3/imagick"
)

func NewHandler(debug bool, dev bool) http.Handler {
	imagick.Initialize()

	config := dims.ReadConfig()

	if debug {
		fmt.Printf("config: %+v\n", config)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/dims4/{clientId}/{signature}/{timestamp}/{commands...}",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDims4(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-sizer/{url}",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsSizer(config, debug, dev, w, r)
		})

	mux.HandleFunc("/dims-status",
		func(w http.ResponseWriter, r *http.Request) {
			dims.HandleDimsStatus(config, debug, dev, w, r)
		})

	return mux
}
