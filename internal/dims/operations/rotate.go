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
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

func RotateCommand(image *vips.ImageRef, args string) error {
	slog.Debug("RotateCommand", "args", args)

	degrees, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	idx, idy, odx, ody := 0.0, 0.0, 0.0, 0.0
	if degrees == 90 {
		idx, idy = 0.0, 1.0
		odx, ody = 1.0, 0.0
	} else if degrees == 180 {
		idx, idy = 0.0, 1.0
		odx, ody = 1.0, 0.0
	} else if degrees == 270 {
		idx, idy = 1.0, 0.0
		odx, ody = 0.0, -1.0
	}

	return image.Similarity(1.0, degrees, &vips.ColorRGBA{}, idx, idy, odx, ody)
}
