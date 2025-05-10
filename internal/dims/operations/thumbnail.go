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

package operations

import (
	"log/slog"
	"math"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func ThumbnailCommand(image *vips.ImageRef, args string) error {
	slog.Debug("LegacyThumbnailCommand", "args", args)

	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	var rect = geometry.ParseGeometry(resizedArgs)

	if err := image.ThumbnailWithSize(int(rect.Width), int(rect.Height), vips.InterestingLow, vips.SizeUp); err != nil {
		return err
	}

	return nil
}

func LegacyThumbnailCommand(image *vips.ImageRef, args string) error {
	slog.Debug("LegacyThumbnailCommand", "args", args)

	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry (with fill)
	var rect = geometry.ParseGeometry(resizedArgs)
	rect = rect.ApplyMeta(image)

	if err := image.Thumbnail(int(rect.Width), int(rect.Height), vips.InterestingAll); err != nil {
		return err
	}

	// Parse Geometry (actual requested size)
	rect = geometry.ParseGeometry(args)

	width := image.Width()
	height := image.Height()
	cropWidth := int(math.Min(rect.Width, float64(width)))
	cropHeight := int(math.Min(rect.Height, float64(height)))

	x := int(math.Max(0, math.Floor(float64(width-cropWidth)/2)))
	y := int(math.Max(0, math.Floor(float64(height-cropHeight)/2)))

	// Clamp if crop area would exceed image bounds
	if x+cropWidth > width || cropWidth == 0 {
		cropWidth = width - x
	}
	if y+cropHeight > height || cropHeight == 0 {
		cropHeight = height - y
	}

	return image.Crop(x, y, cropWidth, cropHeight)
}
