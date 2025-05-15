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
	"github.com/beetlebugorg/go-dims/internal/core"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

type RequestContext interface {
	Config() core.Config
	Validate() bool
	FetchImage(timeout time.Duration) (*core.Image, error)
	LoadImage(image *core.Image) (*vips.ImageRef, error)
	ProcessImage(img *vips.ImageRef, strip bool) (string, []byte, error)
	SendImage(status int, imageFormat string, imageBlob []byte) error
}

func Handler(request RequestContext) error {
	// Validate the request.
	if !request.Validate() {
		return core.NewStatusError(403, "Invalid signature")
	}

	// Download image.
	timeout := time.Duration(request.Config().Timeout.Download) * time.Millisecond
	sourceImage, err := request.FetchImage(timeout)
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
	if err := request.SendImage(200, imageType, imageBlob); err != nil {
		return err
	}

	return nil
}
