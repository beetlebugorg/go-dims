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

package v4

import (
	"errors"
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func ThumbnailCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("ThumbnailCommand", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo
	var exception imagick.ExceptionInfo

	slog.Info("ThumbnailCommand", "resizedArgs", resizedArgs)

	imagick.SetGeometry(mw.GetImageFromMagickWand(), &rect)
	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	slog.Debug("ThumbnailCommand[resize]", "rect", rect)

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		if err := mw.SetSamplingFactors(factors); err != nil {
			return err
		}
	}

	if err := mw.ThumbnailImage(rect.Width, rect.Height); err != nil {
		return err
	}

	if (flags & imagick.PERCENTVALUE) != 0 {
		flags = imagick.ParseGravityGeometry(mw.GetImageFromMagickWand(), args, &rect, &exception)
		if (flags & imagick.ALLVALUES) == 0 {
			return errors.New("parsing thumbnail (crop) geometry failed")
		}

		slog.Debug("ThumbnailCommand[crop]", "rect", rect)
		if err := mw.CropImage(rect.Width, rect.Height, rect.X, rect.Y); err != nil {
			return err
		}

		return mw.SetImagePage(rect.Width, rect.Height, rect.X, rect.Y)
	}

	return nil
}
