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
	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func ThumbnailCommand(image *vips.ImageRef, args string) error {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return err
	}

	if rect.Flags.Force {
		return ResizeCommand(image, args)
	}

	return image.Thumbnail(int(rect.Width), int(rect.Height), vips.InterestingLow)
}

func LegacyThumbnailCommand(image *vips.ImageRef, args string) error {
	rect, err := geometry.ParseGeometry(args)
	if err != nil {
		return err
	}

	return image.Thumbnail(int(rect.Width), int(rect.Height), vips.InterestingCentre)
}
