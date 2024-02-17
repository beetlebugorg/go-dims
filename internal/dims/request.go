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
	"gopkg.in/gographics/imagick.v3/imagick"
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

type Kernel interface {
	ValidateSignature() bool
	FetchImage() error
	ProcessImage() (string, []byte, error)
	SendHeaders(w http.ResponseWriter)
	SendImage(w http.ResponseWriter, status int, imageType string, imageBlob []byte) error
	SendError(w http.ResponseWriter, status int, message string)
}

type Request struct {
	Id                     string    // The hash of the request -> hash(clientId + commands + imageUrl).
	Signature              string    // The signature of the request.
	Config                 Config    // The global configuration.
	ClientId               string    // The client ID of this request.
	ImageUrl               string    // The image URL that is being manipulated.
	SendContentDisposition bool      // The content disposition of the request.
	Commands               []Command // The commands (resize, crop, etc).
	Error                  bool      // Whether the error image is being served.
	SourceImage            Image     // The source image.
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
	if _, err := url.ParseRequestURI(imageUrl); err != nil {
		return Image{
			Status: 400,
		}
	}

	http.DefaultClient.Timeout = timeout
	image, err := http.Get(imageUrl)
	if err != nil || image.StatusCode != 200 {
		return Image{
			Status: image.StatusCode,
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

/*
Parse through the requested commands and set
the optimal image size on the MagicWand.

This is used while reading an image to improve
performance when generating thumbnails from very
large images.

An example speed is taking 1817x3000 sized image and
reducing it to a 78x110 thumbnail:

	without MagickSetSize: 396ms
	with MagickSetSize:    105ms
*/
func (r *Request) setOptimalImageSize(mw *imagick.MagickWand) {
	for _, command := range r.Commands {
		if command.Name == "thumbnail" || command.Name == "resize" {
			var rect imagick.RectangleInfo
			flags := imagick.ParseAbsoluteGeometry(command.Args, &rect)

			if (flags&imagick.WIDTHVALUE != 0) &&
				(flags&imagick.HEIGHTVALUE != 0) &&
				(flags&imagick.PERCENTVALUE == 0) {

				mw.SetSize(rect.Width, rect.Height)

				return
			}
		}
	}
}

// ProcessImage will execute the commands on the image.
func (r *Request) ProcessImage() (string, []byte, error) {
	slog.Debug("executeImagemagick")

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image.
	r.setOptimalImageSize(mw)
	err := mw.ReadImageBlob(r.SourceImage.Bytes)
	if err != nil {
		return "", nil, err
	}

	// Convert image to RGB from CMYK.
	if mw.GetImageColorspace() == imagick.COLORSPACE_CMYK {
		profiles := mw.GetImageProfiles("icc")
		if profiles != nil {
			mw.ProfileImage("ICC", CmykIccProfile)
		}
		mw.ProfileImage("ICC", RgbIccProfile)

		err = mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
		if err != nil {
			return "", nil, err
		}
	}

	// Flip image orientation, if needed.
	mw.AutoOrientImage()

	// Execute the commands.
	stripMetadata := true
	formatProvided := false

	for _, command := range r.Commands {
		if command.Name == "strip" {
			stripMetadata = false
		}

		if command.Name == "format" {
			formatProvided = true
		}

		if err := command.Operation(mw, command.Args); err != nil {
			return "", nil, err
		}

		mw.MergeImageLayers(imagick.IMAGE_LAYER_TRIM_BOUNDS)
	}

	// Strip metadata. (if not already stripped)
	if stripMetadata && r.Config.StripMetadata {
		mw.StripImage()
	}

	// Set output format if not provided in the request.
	if !formatProvided && r.Config.OutputFormat.OutputFormat != "" {
		format := strings.ToLower(mw.GetImageFormat())
		if !contains(r.Config.OutputFormat.Exclude, format) {
			mw.SetImageFormat(r.Config.OutputFormat.OutputFormat)
		}
	}

	mw.ResetIterator()

	return mw.GetImageFormat(), mw.GetImagesBlob(), nil
}

func (r *Request) SendHeaders(w http.ResponseWriter) {
	maxAge := r.Config.OriginCacheControl.Default
	edgeControlTtl := r.Config.EdgeControl.DownstreamTtl

	if r.Config.OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			min := r.Config.OriginCacheControl.Min
			if min != 0 && maxAge <= min {
				maxAge = min
			}

			// If above maximum, set to maximum.
			max := r.Config.OriginCacheControl.Max
			if max != 0 && maxAge >= max {
				maxAge = max
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
		if r.Config.Signing.SigningAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		io.WriteString(h, r.Id)
		if r.SourceImage.Etag != "" {
			io.WriteString(h, r.SourceImage.Etag)
		} else if r.SourceImage.LastModified != "" {
			io.WriteString(h, r.SourceImage.LastModified)
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
	w.Write(imageBlob)

	return nil
}

func (r *Request) SendError(w http.ResponseWriter, status int, message string) {
	slog.Info("sendError", "status", status, "message", message)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	pw := imagick.NewPixelWand()
	defer pw.Destroy()

	pw.SetColor(r.Config.Error.Background)

	mw.NewImage(1, 1, pw)

	// Execute the commands on the placeholder image, giving it the same dimensions as the requested image.
	for _, command := range r.Commands {
		if command.Name == "crop" || command.Name == "thumbnail" {
			var rect imagick.RectangleInfo
			imagick.ParseAbsoluteGeometry(command.Args, &rect)

			if rect.Width > 0 && rect.Height == 0 {
				command.Args = fmt.Sprintf("%d", rect.Height)
			} else if rect.Height > 0 && rect.Width == 0 {
				command.Args = fmt.Sprintf("x%d", rect.Width)
			} else if rect.Width > 0 && rect.Height > 0 {
				command.Args = fmt.Sprintf("%dx%d", rect.Width, rect.Height)
			}

			command.Name = "resize"
			command.Operation = func(mw *imagick.MagickWand, args string) error {
				var rect imagick.RectangleInfo
				imagick.SetGeometry(mw.Image(), &rect)
				flags := imagick.ParseMetaGeometry(args, &rect.X, &rect.Y, &rect.Width, &rect.Height)
				if (flags & imagick.ALLVALUES) != 0 {
					return mw.ScaleImage(rect.Width, rect.Height)
				}

				return nil
			}
		}

		if err := command.Operation(mw, command.Args); err != nil {
			return
		}
	}

	// Set the format for the error image.
	format := mw.GetImageFormat()
	if format == "" && r.Config.OutputFormat.OutputFormat != "" {
		format = r.Config.OutputFormat.OutputFormat
	} else if format == "" {
		format = "png"
	}
	mw.SetImageFormat(format)

	r.Error = true
	err := r.SendImage(w, status, format, mw.GetImageBlob())
	if err != nil {
		slog.Error("sendError failed.", "error", err)
	}
}

func sourceMaxAge(header string) (int, error) {
	if header == "" {
		return 0, errors.New("empty header")
	}

	regex, _ := regexp.Compile("max-age=([\\d]+)")
	match := regex.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0, errors.New("unable to convert to int")
		}

		return sourceMaxAge, nil
	}

	return 0, errors.New("max-age not found in header")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
