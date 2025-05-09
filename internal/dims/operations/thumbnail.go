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
	"errors"
	"log/slog"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func ThumbnailCommand(image *vips.ImageRef, args string) error {
	slog.Debug("ThumbnailCommand", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect = geometry.ParseGeometry(resizedArgs)

	width := image.Width()
	height := image.Height()

	rect.X = image.OffsetX()
	rect.Y = image.OffsetY()
	rect.Width = float64(width)
	rect.Height = float64(height)

	// Parse Meta Geometry
	rect = rect.ApplyMeta(image)
	if rect.Width == 0 || rect.Height == 0 {
		return errors.New("invalid geometry")
	}

	slog.Debug("ThumbnailCommand[resize]", "rect", rect)

	if err := image.Thumbnail(width, height, vips.InterestingAll); err != nil {
		return err
	}

	return nil
}

func LegacyThumbnailCommand(image *vips.ImageRef, args string) error {
	return nil
}
