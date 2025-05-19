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
