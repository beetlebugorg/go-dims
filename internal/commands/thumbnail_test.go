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
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

// Expected: Resize to 256x256, no cropping.
func TestThumbnail(t *testing.T) {
	path := "grid.png"
	args := "256x256"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, img.Width(), 256)
			assert.Equal(t, img.Height(), 256)
		},
		nil, // use default ExportNative
	)
}

// Expected: Resize to 256x256, then crop to 256x100+0+0.
func TestThumbnailWithCrop(t *testing.T) {
	path := "grid.png"
	args := "256x128"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, img.Width(), 256)
			assert.Equal(t, img.Height(), 128)
		},
		nil, // use default ExportNative
	)
}

// Expected: Resize with aspect ratio preserved.
func TestThumbnailWithoutHeight(t *testing.T) {
	path := "pexels-photo-1539116.jpeg"
	args := "32x"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 32, img.Width())
			assert.Equal(t, 40, img.Height())
		},
		nil, // use default ExportNative
	)
}

// Expected: Resize to 256x128, ignoring aspect ratio. No cropping.
func TestThumbnailIgnoreAspectRatio(t *testing.T) {
	path := "grid.png"
	args := "256x128!"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, img.Width(), 256)
			assert.Equal(t, img.Height(), 128)
		},
		nil, // use default ExportNative
	)
}

// Expected: Resize to 256x128, cropping in the center.
func TestLegacyThumbnail(t *testing.T) {
	path := "grid.png"
	args := "256x128"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return LegacyThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 256, img.Width())
			assert.Equal(t, 128, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestLegacyThumbnailWithoutHeight(t *testing.T) {
	path := "pexels-photo-1539116.jpeg"
	args := "32x"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return LegacyThumbnailCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 32, img.Width())
			assert.Equal(t, 20, img.Height())
		},
		nil, // use default ExportNative
	)
}
