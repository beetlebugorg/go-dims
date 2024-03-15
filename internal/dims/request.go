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
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Id                     string // The hash of the request -> hash(clientId + commands + imageUrl).
	Signature              string // The signature of the request.
	Config                 Config // The global configuration.
	ClientId               string // The client ID of this request.
	ImageUrl               string // The image URL that is being manipulated.
	SendContentDisposition bool   // The content disposition of the request.
	RawCommands            string // The commands ('resize/100x100', 'strip/true/format/png', etc).
	Error                  bool   // Whether the error image is being served.
	SourceImage            Image  // The source image.
}

// FetchImage downloads the image from the given URL.
func (r *Request) FetchImage() error {
	slog.Info("downloadImage", "url", r.ImageUrl)

	timeout := time.Duration(r.Config.Timeout.Download) * time.Millisecond
	r.SourceImage = _fetchImage(r.ImageUrl, timeout)

	if r.SourceImage.Status != 200 {
		return fmt.Errorf("failed to download image")
	}

	return nil
}

func _fetchImage(imageUrl string, timeout time.Duration) Image {
	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return Image{
			Status: 400,
		}
	}

	http.DefaultClient.Timeout = timeout
	image, err := http.Get(imageUrl)
	if err != nil {
		return Image{
			Status: 500,
		}
	}

	sourceImage := Image{
		Status:       image.StatusCode,
		EdgeControl:  image.Header.Get("Edge-Control"),
		CacheControl: image.Header.Get("Cache-Control"),
		LastModified: image.Header.Get("Last-Modified"),
		Etag:         image.Header.Get("Etag"),
		Format:       image.Header.Get("Content-Type"),
	}

	sourceImage.Size = int(image.ContentLength)
	sourceImage.Bytes, _ = io.ReadAll(image.Body)

	return sourceImage
}

func (r *Request) SendHeaders(w http.ResponseWriter) {
	maxAge := r.Config.OriginCacheControl.Default
	edgeControlTtl := r.Config.EdgeControl.DownstreamTtl

	if r.Config.OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.Config.OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.Config.OriginCacheControl.Max
			if maxCacheAge != 0 && maxAge >= maxCacheAge {
				maxAge = maxCacheAge
			}
		}
	}

	if r.Error {
		maxAge = r.Config.OriginCacheControl.Error
	}

	// Set cache headers.
	if maxAge > 0 {
		slog.Debug("sendImage", "maxAge", maxAge)

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		w.Header().Set("Expires",
			fmt.Sprintf("%s", time.Now().
				Add(time.Duration(maxAge)*time.Second).
				UTC().
				Format(http.TimeFormat)))
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

		slog.Debug("sendImage", "sendContentDisposition", r.SendContentDisposition, "filename", filename)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.SourceImage.Etag != "" || r.SourceImage.LastModified != "" {
		var h hash.Hash
		if r.Config.Signing.EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.Id))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		} else if r.SourceImage.LastModified != "" {
			h.Write([]byte(r.SourceImage.LastModified))
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("Etag", etag)
	}
}

func (r *Request) SendImage(w http.ResponseWriter, status int, imageFormat string, imageBlob []byte) error {
	slog.Info("SendImage", "status", status, "format", imageFormat, "size", len(imageBlob))

	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	w.WriteHeader(status)

	// Write the image.
	_, err := w.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) SendError(w http.ResponseWriter, status int, message string) {
	slog.Info("sendError", "status", status, "message", message)
}

func (r *Request) Commands() []Command {
	commands := make([]Command, 0)
	parsedCommands := strings.Split(strings.Trim(r.RawCommands, "/"), "/")
	for i := 0; i < len(parsedCommands)-1; i += 2 {
		command := parsedCommands[i]
		args := parsedCommands[i+1]

		commands = append(commands, Command{
			Name: command,
			Args: args,
		})

		slog.Debug("parsedCommand", "command", command, "args", args)
	}

	return commands
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
