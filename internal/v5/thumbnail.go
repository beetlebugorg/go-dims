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

package v5

import (
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
	"strings"
)

func ThumbnailCommand(request *RequestV5, args string) error {
	slog.Debug("ThumbnailCommand", "args", args)

	image := request.vipsImage

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo

	rect.X = image.OffsetX()
	rect.Y = image.OffsetY()
	rect.Width = uint(image.Width())
	rect.Height = uint(image.Height())

	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	slog.Debug("ThumbnailCommand[resize]", "rect", rect)

	if err := image.Thumbnail(int(rect.Width), int(rect.Height), vips.InterestingAll); err != nil {
		return err
	}

	return nil
}

func LegacyThumbnailCommand(request *RequestV5, args string) error {
	return nil
}
