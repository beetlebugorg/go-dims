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
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/commands"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"hash"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beetlebugorg/go-dims/internal/gox/imagex/colorx"
	"github.com/davidbyttow/govips/v2/vips"
)

type Request struct {
	dims.Request

	httpRequest  *http.Request
	httpResponse http.ResponseWriter
}

//-- Request/RequestContext Implementation

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	signature := r.PathValue("signature")
	requestUrl := r.URL
	imageUrl := r.URL.Query().Get("url")
	cmds := r.PathValue("commands")

	eurl := r.URL.Query().Get("eurl")
	if eurl != "" {
		decryptedUrl, err := core.DecryptURL(config.SigningKey, eurl)
		if err != nil {
			slog.Error("DecryptURL failed.", "error", err)
			return nil, fmt.Errorf("DecryptURL failed: %w", err)
		}

		imageUrl = decryptedUrl
	}

	// Signed Parameters
	// _keys query parameter is a comma-delimited list of keys to include in the signature.
	var signedParams map[string]string
	params := r.URL.Query().Get("_keys")
	if params != "" {
		keys := strings.Split(params, ",")
		for _, key := range keys {
			value := r.URL.Query().Get(key)
			if value != "" {
				signedParams[key] = value
			}
		}
	}

	return &Request{
		Request:      *dims.NewRequest(requestUrl, imageUrl, cmds, signedParams, signature, config),
		httpRequest:  r,
		httpResponse: w,
	}, nil
}

func (r *Request) HashId() string {
	h := md5.New()
	h.Write([]byte(r.RawCommands))
	h.Write([]byte(r.ImageUrl))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Request) SendHeaders() {
	w := r.httpResponse

	maxAge := r.Config().OriginCacheControl.Default
	edgeControlTtl := r.Config().EdgeControl.DownstreamTtl

	if r.Config().OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.Config().OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.Config().OriginCacheControl.Max
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
		if r.Config().EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.HashId()))
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

func (r *Request) SendImage(status int, imageFormat string, imageBlob []byte) error {
	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	r.SendHeaders()

	// Set content type.
	r.httpResponse.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	r.httpResponse.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	r.httpResponse.WriteHeader(status)

	// Write the image.
	_, err := r.httpResponse.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) SendError(err error) error {
	message := err.Error()

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var statusError *core.StatusError
	var operationError *commands.OperationError
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

	backgroundColor, err := colorx.ParseHexColor(r.Config().Error.Background)
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

	r.SourceImage = core.Image{
		Status: status,
		Format: vips.ImageTypes[vips.ImageTypeJPEG],
	}

	// Send error headers.
	maxAge := r.Config().OriginCacheControl.Error
	if maxAge > 0 {
		r.httpResponse.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		r.httpResponse.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
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

func sourceMaxAge(header string) (int, error) {
	if header == "" {
		return 0, errors.New("empty header")
	}

	pattern, _ := regexp.Compile(`max-age=(\d+)`)
	match := pattern.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0, errors.New("unable to convert to int")
		}

		return sourceMaxAge, nil
	}

	return 0, errors.New("max-age not found in header")
}
