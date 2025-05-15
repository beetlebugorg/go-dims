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

package commands

import (
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func ThumbnailCommand(image *vips.ImageRef, args string) error {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return NewOperationError("thumbnail", args, err.Error())
	}

	cropMethod := vips.InterestingLow
	if rect.Height == 0 {
		cropMethod = vips.InterestingNone
		rect.Height = 99999
	}

	if rect.Flags.Force {
		return ResizeCommand(image, args)
	}

	err = image.Thumbnail(int(rect.Width), int(rect.Height), cropMethod)
	if err != nil {
		return NewOperationError("thumbnail", args, err.Error())
	}

	return nil
}

func LegacyThumbnailCommand(image *vips.ImageRef, args string) error {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return NewOperationError("thumbnail", args, err.Error())
	}

	cropMethod := vips.InterestingCentre
	if rect.Height == 0 {
		cropMethod = vips.InterestingHigh

		cropRect := rect.ApplyMeta(image)
		rect.Height = cropRect.Height / 2
	}

	err = image.Thumbnail(int(rect.Width), int(rect.Height), cropMethod)
	if err != nil {
		return NewOperationError("thumbnail", args, err.Error())
	}

	return nil
}
