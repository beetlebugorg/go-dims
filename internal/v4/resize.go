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
)

func ResizeCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("ResizeCommand", "args", args)

	// Parse Geometry
	var rect imagick.RectangleInfo

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(args, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	slog.Debug("ResizeCommand", "width", rect.Width, "height", rect.Height)

	return mw.ScaleImage(rect.Width, rect.Height)
}
