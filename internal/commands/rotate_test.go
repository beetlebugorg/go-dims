package commands

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
