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
	"net/http"
	"time"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

/*
type RequestHandler interface {
	FetchImage(timeout time.Duration) ([]byte, error)
	LoadImage(sourceImage []byte) (*core.VipsImage, error)
	ProcessImage(vipsImage *core.VipsImage, isLambda bool) (string, []byte, error)
	SendHeaders(w http.ResponseWriter)
	SendImage(w http.ResponseWriter, statusCode int, imageType string, imageBlob []byte) error
}
*/

func Handler(request Request, config core.Config, w http.ResponseWriter) error {

	// Download image.
	var fi core.ImageFetcher = &request
	timeout := time.Duration(request.Config.Timeout.Download) * time.Millisecond
	sourceImage, err := fi.FetchImage(timeout)
	if err != nil {
		return err
	}

	// Convert image to vips image.
	vipsImage, err := request.LoadImage(sourceImage)
	if err != nil {
		return err
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := request.ProcessImage(vipsImage, false)
	if err != nil {
		return err
	}

	// Serve the image.
	request.SendHeaders(w)
	if err := request.SendImage(w, 200, imageType, imageBlob); err != nil {
		return err
	}

	return nil
}
