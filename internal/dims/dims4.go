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
	"bytes"
	"crypto/md5"
	"crypto/sha256"
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

	"gopkg.in/gographics/imagick.v3/imagick"
)

type request struct {
	config                 *Config     // The global configuration.
	clientId               string      // The client ID of this request.
	signature              string      // The signature of the request.
	timestamp              int32       // The timestamp of the request.
	imageUrl               string      // The image URL that is being manipulated.
	sendContentDisposition bool        // The content disposition of the request.
	commands               string      // The unparsed commands (resize, crop, etc).
	requestHash            string      // The hash of the request.
	sampleFactor           float64     // The sample factor for optimizing resizing.
	placeholder            bool        // Whether the placeholder image is being served.
	sourceImage            sourceImage // The source image.
}

type sourceImage struct {
	originalImage       []byte // The downloaded image.
	originalImageSize   int    // The original image size in bytes.
	originalImageFormat string // The original image format.
	status              int    // The HTTP status code of the downloaded image.
	cacheControl        string // The cache headers from the downloaded image.
	edgeControl         string // The edge control headers from the downloaded image.
	lastModified        string // The last modified header from the downloaded image.
	etag                string // The etag header from the downloaded image.
}

func HandleDims4(config Config, debug bool, dev bool, w http.ResponseWriter, r *http.Request) {
	slog.Info("handleDims5()",
		"imageUrl", r.URL.Query().Get("url"),
		"clientId", r.PathValue("clientId"),
		"signature", r.PathValue("signature"),
		"timestamp", r.PathValue("timestamp"),
		"commands", r.PathValue("commands"))

	request := newRequest(r, &config)

	// Verify signature.
	if !dev {
		if err := request.verifySignature(); err != nil {
			request.sourceImage = sourceImage{
				status: 500,
			}
			request.sendPlaceholderImage(w)

			return
		}
	}

	// Download image.
	if err := request.fetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		request.sendPlaceholderImage(w)

		return
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := request.processImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		request.sendPlaceholderImage(w)

		return
	}

	// Serve the image.
	if err := request.sendImage(w, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}

func newRequest(r *http.Request, config *Config) request {
	var timestamp int32
	fmt.Sscanf(r.PathValue("timestamp"), "%d", &timestamp)

	var h hash.Hash
	if config.SigningAlgorithm == "md5" {
		h = md5.New()
	} else {
		h = sha256.New()
	}

	io.WriteString(h, r.PathValue("clientId"))
	io.WriteString(h, r.PathValue("commands"))
	io.WriteString(h, r.URL.Query().Get("url"))
	requestHash := fmt.Sprintf("%x", h.Sum(nil))

	return request{
		config:      config,
		clientId:    r.PathValue("clientId"),
		imageUrl:    r.URL.Query().Get("url"),
		timestamp:   timestamp,
		commands:    r.PathValue("commands"),
		requestHash: requestHash,
		signature:   r.PathValue("signature"),
	}
}

// verifySignature verifies the signature of the image resize is valid.
func (r *request) verifySignature() error {
	slog.Debug("verifySignature", "url", r.imageUrl)

	var h string
	if r.config.SigningAlgorithm == "md5" {
		h = Sign(fmt.Sprintf("%d", r.timestamp), r.config.SecretKey, r.commands, r.imageUrl)
	} else {
		h = SignHmacSha256(fmt.Sprintf("%d", r.timestamp), r.config.SecretKey, r.commands, r.imageUrl)
	}

	if !bytes.Equal([]byte(h), []byte(r.signature)) {
		slog.Error("verifySignature failed.", "expected", h, "got", r.signature)

		return fmt.Errorf("invalid signature")
	}

	return nil
}

func _fetchImage(url string, timeout time.Duration) sourceImage {
	http.DefaultClient.Timeout = timeout
	image, err := http.Get(url)
	if err != nil || image.StatusCode != 200 {
		return sourceImage{
			status: image.StatusCode,
		}
	}

	sourceImage := sourceImage{
		status:              image.StatusCode,
		edgeControl:         image.Header.Get("Edge-Control"),
		cacheControl:        image.Header.Get("Cache-Control"),
		lastModified:        image.Header.Get("Last-Modified"),
		etag:                image.Header.Get("Etag"),
		originalImageFormat: image.Header.Get("Content-Type"),
	}

	sourceImage.originalImageSize = int(image.ContentLength)
	sourceImage.originalImage, _ = io.ReadAll(image.Body)

	return sourceImage
}

func (r *request) fetchImage() error {
	slog.Info("downloadImage", "url", r.imageUrl)

	timeout := time.Duration(r.config.DownloadTimeout) * time.Millisecond
	r.sourceImage = _fetchImage(r.imageUrl, timeout)

	if r.sourceImage.status != 200 {
		return fmt.Errorf("failed to download image")
	}

	return nil
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
func (r *request) setOptimalImageSize(mw *imagick.MagickWand) {
	explodedCommands := strings.Split(r.commands, "/")
	for i := 0; i < len(explodedCommands)-1; i += 2 {
		command := explodedCommands[i]
		args := explodedCommands[i+1]

		if command == "thumbnail" || command == "resize" {
			var rect imagick.RectangleInfo
			flags := imagick.ParseAbsoluteGeometry(args, &rect)

			if (flags&imagick.WIDTHVALUE != 0) &&
				(flags&imagick.HEIGHTVALUE != 0) &&
				(flags&imagick.PERCENTVALUE == 0) {
				mw.SetSize(rect.Width, rect.Height)
				return
			}
		}
	}
}

/*
This is the main code for processing images.  It will parse
the command string into individual commands and execute them.

When it's finished it will write the content type header and
image data to connection and flush the connection.

Commands should always come in pairs, the command name followed
by the commands arguments delimited by '/'.  Example:

	thumbnail/78x110/quality/70

This will first execute the thumbnail command then it will
set the quality of the image to 70 before writing the image
to the connection.
*/
func (r *request) processImage() (string, []byte, error) {
	slog.Debug("executeImagemagick", "commands", r.commands)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image.
	r.setOptimalImageSize(mw)
	err := mw.ReadImageBlob(r.sourceImage.originalImage)
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

	explodedCommands := strings.Split(r.commands, "/")
	for i := 0; i < len(explodedCommands)-1; i += 2 {
		command := explodedCommands[i]
		args := explodedCommands[i+1]

		if command == "strip" {
			stripMetadata = false
		}

		if command == "format" {
			formatProvided = true
		}

		// Lookup command, call it.
		if operation, ok := Operations[command]; ok {
			if err := operation(mw, args); err != nil {
				return "", nil, err
			}

			mw.MergeImageLayers(imagick.IMAGE_LAYER_TRIM_BOUNDS)
		}
	}

	// Strip metadata. (if not already stripped)
	if stripMetadata && r.config.StripMetadata {
		mw.StripImage()
	}

	// Set output format if not provided in the request.
	if !formatProvided && r.config.DefaultOutputFormat != "" {
		format := strings.ToLower(mw.GetImageFormat())
		if !contains(r.config.IgnoreDefaultOutputFormats, format) {
			mw.SetImageFormat(r.config.DefaultOutputFormat)
		}
	}

	mw.ResetIterator()

	return mw.GetImageFormat(), mw.GetImagesBlob(), nil
}

func (r *request) sendPlaceholderImage(w http.ResponseWriter) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	pw := imagick.NewPixelWand()
	defer pw.Destroy()

	pw.SetColor(r.config.PlaceholderBackground)

	mw.NewImage(1, 1, pw)

	// Execute the commands on the placeholder image, giving it the same dimensions as the requested image.
	explodedCommands := strings.Split(r.commands, "/")
	for i := 0; i < len(explodedCommands)-1; i += 2 {
		command := explodedCommands[i]
		args := explodedCommands[i+1]

		if command == "crop" || command == "thumbnail" {
			var rect imagick.RectangleInfo
			imagick.ParseAbsoluteGeometry(args, &rect)

			if rect.Width > 0 && rect.Height == 0 {
				args = fmt.Sprintf("%d", rect.Height)
			} else if rect.Height > 0 && rect.Width == 0 {
				args = fmt.Sprintf("x%d", rect.Width)
			} else if rect.Width > 0 && rect.Height > 0 {
				args = fmt.Sprintf("%dx%d", rect.Width, rect.Height)
			}

			command = "resize"
		}

		// Lookup command, call it.
		if operation, ok := Operations[command]; ok {
			operation(mw, args)
		}
	}

	r.placeholder = true
	r.sendImage(w, mw.GetImageFormat(), mw.GetImageBlob())
}

func (r *request) sendImage(w http.ResponseWriter, imageType string, imageBlob []byte) error {
	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	cacheControlMaxAge := r.config.CacheControlMaxAge
	edgeControlTtl := r.config.EdgeControlDownstreamTtl

	sourceMaxAge := sourceMaxAge(r.sourceImage.cacheControl)
	trustSourceImage := false

	// Do we trust the source image (i.e. we control the origin) and are we able to pull out
	// the max-age from its Cache-Control header?
	if r.config.TrustSrc && sourceMaxAge > 0 {

		// Do we have valid min and max cache control values?
		if r.config.MinSrcCacheControl >= -1 && r.config.MaxSrcCacheControl >= -1 {

			// Is the max-age value between the min and max? Use the source image value.
			if (sourceMaxAge >= r.config.MinSrcCacheControl || r.config.MinSrcCacheControl == -1) &&
				(sourceMaxAge <= r.config.MaxSrcCacheControl || r.config.MaxSrcCacheControl == -1) {
				trustSourceImage = true
			}
		}

		if trustSourceImage {
			cacheControlMaxAge = sourceMaxAge
		}
	}

	if r.placeholder {
		cacheControlMaxAge = r.config.PlaceholderImageExpire
	}

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageType)))

	// Set cache headers.
	if cacheControlMaxAge > 0 {
		slog.Debug("sendImage", "cacheControlMaxAge", cacheControlMaxAge)

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", cacheControlMaxAge))
		w.Header().Set("Expires",
			fmt.Sprintf("%s", time.Now().
				Add(time.Duration(cacheControlMaxAge)*time.Second).
				UTC().
				Format(http.TimeFormat)))
	}

	if edgeControlTtl > 0 {
		w.Header().Set("Edge-Control", fmt.Sprintf("downstream-ttl=%d, public", edgeControlTtl))
	}

	// Set content disposition.
	if r.sendContentDisposition {
		// Grab filename from imageUrl
		u, err := url.Parse(r.imageUrl)
		if err != nil {
			return err
		}

		filename := filepath.Base(u.Path)

		slog.Debug("sendImage", "sendContentDisposition", r.sendContentDisposition, "filename", filename)

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	}

	// Set etag header.
	if r.sourceImage.etag != "" || r.sourceImage.lastModified != "" {
		var h hash.Hash
		if r.config.SigningAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		io.WriteString(h, r.requestHash)
		if r.sourceImage.etag != "" {
			io.WriteString(h, r.sourceImage.etag)
		} else if r.sourceImage.lastModified != "" {
			io.WriteString(h, r.sourceImage.lastModified)
		}

		etag := fmt.Sprintf("%x", h.Sum(nil))

		w.Header().Set("Etag", etag)
	}

	// Set content length
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	w.WriteHeader(r.sourceImage.status)

	// Write the image.
	w.Write(imageBlob)

	return nil
}

func sourceMaxAge(header string) int {
	if header == "" {
		return 0
	}

	regex, _ := regexp.Compile("max-age=([\\d]+)")
	match := regex.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0
		}

		return sourceMaxAge
	}

	return 0
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
