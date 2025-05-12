package operations

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResize(t *testing.T) {
	path := "grid.png"
	args := "256x256"

	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	err = ResizeCommand(image, args)
	require.NoError(t, err, "failed to resize image: %s", path)

	assert.Equal(t, image.Width(), 256)
	assert.Equal(t, image.Height(), 256)
}

func TestResizeOnlySmaller(t *testing.T) {
	path := "grid.png"
	args := "256x256<"

	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	err = ResizeCommand(image, args)
	require.NoError(t, err, "failed to resize image: %s", path)

	assert.Equal(t, image.Width(), 512)
	assert.Equal(t, image.Height(), 512)
}

func TestResizeOnlyLarger(t *testing.T) {
	path := "grid.png"
	args := "256x256>"

	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	err = ResizeCommand(image, args)
	require.NoError(t, err, "failed to resize image: %s", path)

	assert.Equal(t, image.Width(), 256)
	assert.Equal(t, image.Height(), 256)
}

func TestResizeMaintainAspectRatio(t *testing.T) {
	path := "grid.png"
	args := "100x50"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ResizeCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 50, img.Width())
			assert.Equal(t, 50, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestResizeIgnoreAspectRatio(t *testing.T) {
	path := "grid.png"
	args := "100x50!"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ResizeCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 100, img.Width())
			assert.Equal(t, 50, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestResizeIgnoreAspectRatioWidthOnly(t *testing.T) {
	path := "grid.png"
	args := "100x!"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ResizeCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 100, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestResizeFill(t *testing.T) {
	path := "grid.png"
	args := "50x100^"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return ResizeCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 100, img.Width())
			assert.Equal(t, 100, img.Height())
		},
		nil, // use default ExportNative
	)
}

func TestResizeFillWithCrop(t *testing.T) {
	path := "grid.png"
	args := "50x100^"
	cropsArgs := "50x"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			err := ResizeCommand(img, args)
			require.NoError(t, err, "failed to resize image: %s", path)

			return CropCommand(img, cropsArgs)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 50, img.Width())
			assert.Equal(t, 100, img.Height())
		},
		nil, // use default ExportNative
	)
}
