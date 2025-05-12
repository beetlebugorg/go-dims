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

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func SharpenCommand(image *vips.ImageRef, args string) error {
	slog.Debug("SharpenCommand", "args", args)

	geo, err := geometry.ParseGeometry(args)
	if err != nil {
		return err
	}

	x1 := geo.Width
	m2 := geo.Height * 2
	if m2 == 0 {
		m2 = 2.0
	}

	slog.Debug("SharpenCommand", "m2", m2, "x1", x1)

	return image.Sharpen(float64(0.5), x1, m2)
}
