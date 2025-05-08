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
	"github.com/davidbyttow/govips/v2/vips"
)

func AutolevelCommand(image *vips.ImageRef, args string) error {
	if args != "true" {
		return nil
	}

	statsOut, err := image.Copy()
	if err != nil {
		return err
	}

	if err := statsOut.Stats(); err != nil {
		return err
	}

	stat, _ := statsOut.GetPoint(0, 0)
	min := stat[0]

	stat, _ = statsOut.GetPoint(1, 0)
	max := stat[0]

	// Compute scale and offset to stretch to [0, 255]
	scale := 255.0 / (max - min)
	offset := -min * scale

	scales := make([]float64, image.Bands())
	offsets := make([]float64, image.Bands())
	for i := range scales {
		scales[i] = scale
		offsets[i] = offset
	}

	// Apply the linear stretch
	if err := image.Linear(scales, offsets); err != nil {
		return err
	}

	return nil
}
