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
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func ResizeCommand(image *vips.ImageRef, args string) error {
	slog.Debug("ResizeCommand", "args", args)

	// Parse Geometry
	rect := imagick.RectangleInfo{
		Width:  uint(image.Width()),
		Height: uint(image.Height()),
		X:      image.OffsetX(),
		Y:      image.OffsetY(),
	}

	flags := imagick.ParseMetaGeometry(args, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	slog.Debug("ResizeCommand", "width", rect.Width, "height", rect.Height)

	xr := float64(rect.Width) / float64(image.Width())
	yr := float64(rect.Height) / float64(image.Height())

	return image.ResizeWithVScale(xr, yr, vips.KernelLanczos3)
}
