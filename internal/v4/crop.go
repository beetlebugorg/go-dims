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

func CropCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("CropCommand", "args", args)

	/* Replace spaces with '+'. This happens when some user agents inadvertently
	 * escape the '+' as %20 which gets converted to a space.
	 *
	 * Example:
	 *
	 * 900x900%20350%200 is '900x900 350 0' which is an invalid, the following code
	 * coverts this to '900x900+350+0'.
	 *
	 */
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")

	// Parse Geometry
	var rect imagick.RectangleInfo
	var exception imagick.ExceptionInfo
	flags := imagick.ParseGravityGeometry(mw.Image(), sanitizedArgs, &rect, &exception)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("invalid geometry")
	}

	slog.Debug("CropCommand", "rect", rect)

	if err := mw.CropImage(rect.Width, rect.Height, rect.X, rect.Y); err != nil {
		return err
	}

	return mw.SetImagePage(rect.Width, rect.Height, rect.X, rect.Y)
}
