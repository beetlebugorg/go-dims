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

package commands

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

func Watermark(image *vips.ImageRef, args string, data RequestOperation) error {
	url := data.URL.Query().Get("overlay")
	if url == "" {
		return NewOperationError("watermark", args, "missing required query parameter 'overlay'")
	}

	// Parse args "<opacity>,<size>,<gravity>"
	//   - opacity should be between 0-1
	//   - size should be between 0-1
	//   - gravity can be: n, ne, nw, s, se, sw, w, e, c
	opacity, size, gravity, err := parseWatermarkArgs(args)
	if err != nil {
		return err
	}

	// Download overlay image
	timeout := time.Duration(data.Config.Timeout.Download) * time.Millisecond
	overlayImageSource, err := core.FetchImage(url, timeout)
	if err != nil {
		return NewOperationError("watermark", args, err.Error())
	}

	overlayImage, err := vips.LoadImageFromBuffer(overlayImageSource.Bytes, nil)
	if err != nil {
		return NewOperationError("watermark", args, err.Error())
	}

	// Resize image
	if err := scaleOverlay(image, overlayImage, size); err != nil {
		return NewOperationError("watermark", args, err.Error())
	}

	// Reduce opacity of overlay image
	// Combine images
	reduceOpacity(overlayImage, opacity)
	overlayImage.Gravity(gravity, image.Width(), image.Height())

	image.Composite(overlayImage, vips.BlendModeOver, 0, 0)

	return nil
}

func reduceOpacity(image *vips.ImageRef, opacity float64) error {
	if !image.HasAlpha() {
		if err := image.AddAlpha(); err != nil {
			return err
		}
	}

	// Extra alpha channel
	alpha, err := image.Copy()
	if err != nil {
		return err
	}

	if err := alpha.ExtractBand(alpha.Bands()-1, 1); err != nil {
		return err
	}

	alpha.Linear1(opacity, 0)

	// Add the new alpha channel.
	if err := image.ExtractBand(0, 3); err != nil {
		return err
	}

	if err := image.BandJoin(alpha); err != nil {
		return err
	}

	return nil
}

// scaleOverlay scales the overlay image based on the largest dimension of the base image.
func scaleOverlay(base *vips.ImageRef, overlay *vips.ImageRef, size float64) error {
	originalWidth := float64(base.Width())
	originalHeight := float64(base.Height())

	overlayWidth := float64(overlay.Width())
	overlayHeight := float64(overlay.Height())

	var largestSize float64
	if originalWidth > originalHeight {
		largestSize = originalWidth * size
	} else {
		largestSize = originalHeight * size
	}

	var finalWidth, finalHeight float64
	if overlayWidth > overlayHeight {
		finalWidth = largestSize
		finalHeight = largestSize / (overlayWidth / overlayHeight)
	} else if overlayWidth < overlayHeight {
		finalWidth = largestSize / (overlayHeight / overlayWidth)
		finalHeight = largestSize
	} else {
		finalWidth = largestSize
		finalHeight = largestSize
	}

	scaleX := finalWidth / overlayWidth
	scaleY := finalHeight / overlayHeight
	scale := math.Min(scaleX, scaleY) // Uniform scale preferred for go-vips

	return overlay.Resize(scale, vips.KernelLanczos3)
}

// parseWatermarkArgs parses a string of the form "<opacity>,<size>,<gravity>"
//   - opacity: float between 0.0 and 1.0
//   - size:    float between 0.0 and 1.0
//   - gravity: one of n, ne, nw, s, se, sw, w, e, c
func parseWatermarkArgs(input string) (opacity, size float64, gravity vips.Gravity, err error) {
	parts := strings.Split(input, ",")
	if len(parts) != 3 {
		err = fmt.Errorf("expected 3 comma‑separated values, got %d", len(parts))
		return
	}

	// parse and validate opacity
	if opacity, err = strconv.ParseFloat(parts[0], 64); err != nil {
		err = fmt.Errorf("invalid opacity %q: %w", parts[0], err)
		return
	}
	if opacity < 0 || opacity > 1 {
		err = fmt.Errorf("opacity %f out of range [0.0,1.0]", opacity)
		return
	}

	// parse and validate size
	if size, err = strconv.ParseFloat(parts[1], 64); err != nil {
		err = fmt.Errorf("invalid size %q: %w", parts[1], err)
		return
	}
	if size < 0 || size > 1 {
		err = fmt.Errorf("size %f out of range [0.0,1.0]", size)
		return
	}

	// validate gravity
	gravityStr := parts[2]
	gravity = vips.GravityCentre
	switch gravityStr {
	case "n":
		gravity = vips.GravityNorth
	case "ne":
		gravity = vips.GravityNorthEast
	case "nw":
		gravity = vips.GravityNorthWest
	case "s":
		gravity = vips.GravitySouth
	case "se":
		gravity = vips.GravitySouthEast
	case "sw":
		gravity = vips.GravitySouthWest
	case "w":
		gravity = vips.GravityWest
	case "e":
		gravity = vips.GravityEast
	case "c":
		gravity = vips.GravityCentre
	default:
		err = fmt.Errorf("invalid gravity %q; must be one of n, ne, nw, s, se, sw, w, e, c", gravity)
		return
	}

	return
}
