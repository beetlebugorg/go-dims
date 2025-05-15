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

package request

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	core2 "github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/operations"
	"hash"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beetlebugorg/go-dims/internal/gox/imagex/colorx"
	"github.com/davidbyttow/govips/v2/vips"
)

func (r *HttpDimsRequest) SendHeaders() {
	w := r.response

	maxAge := r.config.OriginCacheControl.Default
	edgeControlTtl := r.config.EdgeControl.DownstreamTtl

	if r.config.OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.config.OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.config.OriginCacheControl.Max
			if maxCacheAge != 0 && maxAge >= maxCacheAge {
				maxAge = maxCacheAge
			}
		}
	}

	// Set cache headers.
	if maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		w.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
	}

	if edgeControlTtl > 0 {
		w.Header().Set("Edge-Control", fmt.Sprintf("downstream-ttl=%d", edgeControlTtl))
	}

	// Set content disposition.
	if r.SendContentDisposition {
		// Grab filename from imageUrl
		u, err := url.Parse(r.ImageUrl)
		if err != nil {
			return
		}

		filename := filepath.Base(u.Path)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.SourceImage.Etag != "" {
		var h hash.Hash
		if r.config.EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("ETag", etag)
	}

	if r.SourceImage.LastModified != "" {
		w.Header().Set("Last-Modified", r.SourceImage.LastModified)
	}
}

func (r *HttpDimsRequest) SendImage(status int, imageFormat string, imageBlob []byte) error {
	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	r.SendHeaders()

	// Set content type.
	r.response.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	r.response.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	r.response.WriteHeader(status)

	// Write the image.
	_, err := r.response.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *HttpDimsRequest) SendError(err error) error {
	message := err.Error()

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var statusError *core2.StatusError
	var operationError *operations.OperationError
	if errors.As(err, &statusError) {
		status = statusError.StatusCode
	} else if errors.As(err, &operationError) {
		status = operationError.StatusCode
	}

	errorImage, err := vips.Black(512, 512)
	if err != nil {
		return err
	}

	if err := errorImage.BandJoinConst([]float64{0, 0}); err != nil {
		return err
	}

	backgroundColor, err := colorx.ParseHexColor(r.config.Error.Background)
	if err != nil {
		return err
	}

	red, green, blue, _ := backgroundColor.RGBA()
	redI := float64(red) / 65535 * 255
	greenI := float64(green) / 65535 * 255
	blueI := float64(blue) / 65535 * 255

	if err := errorImage.Linear([]float64{0, 0, 0}, []float64{redI, greenI, blueI}); err != nil {
		return err
	}

	r.SourceImage = core2.Image{
		Status: status,
		Format: vips.ImageTypes[vips.ImageTypeJPEG],
	}

	// Send error headers.
	maxAge := r.config.OriginCacheControl.Error
	if maxAge > 0 {
		r.response.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		r.response.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
	}

	imageType, imageBlob, err := r.ProcessImage(errorImage, true)
	if err != nil {
		// If processing failed because of a bad command then return the image as-is.
		exportOptions := vips.NewJpegExportParams()
		exportOptions.Quality = 1
		imageBytes, _, _ := errorImage.ExportJpeg(exportOptions)

		return r.SendImage(status, "jpg", imageBytes)
	}

	if imageType == "" {
		imageType = "jpg"
	}

	return r.SendImage(status, imageType, imageBlob)
}
