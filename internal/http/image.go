// Copyright 2025 Jeremy Collins. All rights reserved.
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

package http

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type imageBackend struct {
}

func init() {
	core.RegisterImageBackend(NewImageBackend())
}

func NewImageBackend() core.ImageBackend {
	return imageBackend{}
}

func (backend imageBackend) Name() string {
	return "http"
}

func (backend imageBackend) CanHandle(imageSource string) bool {
	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		return true
	}

	return false
}

func (backend imageBackend) FetchImage(imageUrl string, timeout time.Duration) (*core.Image, error) {
	slog.Debug("downloadImage", "url", imageUrl)

	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", fmt.Sprintf("go-dims/%s", core.Version))

	http.DefaultClient.Timeout = timeout
	image, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	imageSize := int(image.ContentLength)
	imageBytes, err := io.ReadAll(image.Body)
	if err != nil {
		return nil, err
	}

	sourceImage := core.Image{
		Status:       image.StatusCode,
		EdgeControl:  image.Header.Get("Edge-Control"),
		CacheControl: image.Header.Get("Cache-Control"),
		LastModified: image.Header.Get("Last-Modified"),
		Etag:         image.Header.Get("Etag"),
		Format:       vips.ImageTypes[vips.DetermineImageType(imageBytes)],
		Size:         imageSize,
		Bytes:        imageBytes,
	}

	if image.StatusCode != 200 {
		return nil, &core.StatusError{
			Message:    fmt.Sprintf("failed to fetch image from %s", imageUrl),
			StatusCode: image.StatusCode,
		}
	}

	return &sourceImage, nil
}
