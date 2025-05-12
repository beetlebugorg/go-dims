package operations

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expected: Crop out the 2nd quadrant.
func TestCrop(t *testing.T) {
	path := "grid.png"
	args := "256x256+256+0"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return CropCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, img.Width(), 256)
			assert.Equal(t, img.Height(), 256)
		},
		nil, // use default ExportNative
	)
}

// Expected: Crop out the 2nd quadrant.
func TestCropPercent(t *testing.T) {
	path := "grid.png"
	args := "50%x50%+50%+0"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return CropCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 256, img.Width())
			assert.Equal(t, 256, img.Height())
		},
		nil, // use default ExportNative
	)
}

// Expected: Crop out the 4th quadrant.
func TestCropPercentWithAbsolute(t *testing.T) {
	path := "grid.png"
	args := "50%x50%+50%+256"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return CropCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 256, img.Width())
			assert.Equal(t, 256, img.Height())
		},
		nil, // use default ExportNative
	)
}

// Y offset + Height extends beyond bottom edge of image.
// Expected: Crop out the 3rd and 4th quadrants.
func TestCropRegionLargerThanImage(t *testing.T) {
	path := "grid.png"
	args := "512x512+0+256"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return CropCommand(img, args)
		},
		func(img *vips.ImageRef) {
			require.True(t, img.Width() > 0)
			require.True(t, img.Height() > 0)
		},
		nil, // use default ExportNative
	)
}

// X & Y offsets extend beyond the image.
func TestCropXOffsetOutsideImage(t *testing.T) {
	path := "grid.png"
	args := "256x256+768+768"

	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	err = CropCommand(image, args)
	assert.ErrorContains(t, err, "width must be greater than 0")
}

// X & Y offsets extend beyond the image.
func TestCropYOffsetOutsideImage(t *testing.T) {
	path := "grid.png"
	args := "256x256+0+768"

	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	err = CropCommand(image, args)
	assert.ErrorContains(t, err, "height must be greater than 0")
}
