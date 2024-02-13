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
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type sizerRequest struct {
	imageUrl    string
	sourceImage sizerSourceImage
}

type sizerSourceImage struct {
	image     []byte
	imageSize int
}

type sizerResponse struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func HandleDimsSizer(config Config, debug bool, dev bool, w http.ResponseWriter, r *http.Request) {
	request := sizerRequest{
		imageUrl: r.PathValue("url"),
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	request.fetchImage(request.imageUrl)

	// Read the image.
	mw.ReadImageBlob(request.sourceImage.image)

	w.Header().Set("Content-Type", "application/json")

	json, err := json.Marshal(sizerResponse{
		Height: int(mw.GetImageHeight()),
		Width:  int(mw.GetImageWidth()),
	})

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error marshalling JSON"))
	} else {
		w.WriteHeader(200)
		w.Write(json)
	}
}

func (r *sizerRequest) fetchImage(url string) error {
	image, err := http.Get(url)
	if err != nil || image.StatusCode != 200 {
		return errors.New("Failed to fetch image")
	}

	imageSize := int(image.ContentLength)
	imageBytes, _ := io.ReadAll(image.Body)

	r.sourceImage = sizerSourceImage{
		image:     imageBytes,
		imageSize: imageSize,
	}

	return nil
}
