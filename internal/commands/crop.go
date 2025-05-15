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
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

func CropCommand(image *vips.ImageRef, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+") + "!"

	rect, err := geometry.ParseGeometry(sanitizedArgs)
	if err != nil {
		return NewOperationError("crop", args, err.Error())
	}
	rect = rect.ApplyMeta(image)

	height := rect.Y + int(rect.Height)
	if height > image.Height() {
		rect.Height = float64(image.Height()) - float64(rect.Y)
	}

	width := rect.X + int(rect.Width)
	if width > image.Width() {
		rect.Width = float64(image.Width()) - float64(rect.X)
	}

	if rect.Width <= 0 {
		return NewOperationError("crop", args, "width must be greater than 0")
	}

	if rect.Height <= 0 {
		return NewOperationError("crop", args, "height must be greater than 0")
	}

	err = image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
	if err != nil {
		return NewOperationError("crop", args, err.Error())
	}

	return nil
}
