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

package operations

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

// Expected: Rotate to 45 degrees, expanding the image to fit. Extra space is filled with black, or alpha.
func TestRotate45(t *testing.T) {
	path := "grid.png"
	args := "45"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return RotateCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 724, img.Width())
			assert.Equal(t, 724, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestRotate90(t *testing.T) {
	path := "grid.png"
	args := "90"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return RotateCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 512, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestRotate180(t *testing.T) {
	path := "grid.png"
	args := "180"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return RotateCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 512, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestRotate270(t *testing.T) {
	path := "grid.png"
	args := "270"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return RotateCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 512, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil, // use default ExportNative
	)
}
