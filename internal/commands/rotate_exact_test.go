package commands

import (
	"bytes"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

func loadGrid(t *testing.T) *vips.ImageRef {
	t.Helper()

	vips.Startup(nil)

	image, err := vips.NewImageFromFile(sourceImageDir + "grid.png")
	require.NoError(t, err)

	return image
}

func export(t *testing.T, image *vips.ImageRef) []byte {
	t.Helper()

	buf, _, err := image.ExportNative()
	require.NoError(t, err)

	return buf
}

// A quarter turn is a pure remapping of pixels, so it must be exact. Rotating
// by 180 degrees has to give the same image as flipping both axes. Similarity
// resampled instead, which produced something close but not equal.
func TestRotate180MatchesDoubleFlip(t *testing.T) {
	reference := loadGrid(t)
	defer reference.Close()
	require.NoError(t, reference.Flip(vips.DirectionHorizontal))
	require.NoError(t, reference.Flip(vips.DirectionVertical))

	rotated := loadGrid(t)
	defer rotated.Close()
	require.NoError(t, RotateCommand(rotated, "180"))

	require.True(t, bytes.Equal(export(t, reference), export(t, rotated)),
		"rotate/180 must equal a horizontal then vertical flip")
}

// Two quarter turns must land exactly on a half turn.
func TestRotate90TwiceEquals180(t *testing.T) {
	twice := loadGrid(t)
	defer twice.Close()
	require.NoError(t, RotateCommand(twice, "90"))
	require.NoError(t, RotateCommand(twice, "90"))

	once := loadGrid(t)
	defer once.Close()
	require.NoError(t, RotateCommand(once, "180"))

	require.True(t, bytes.Equal(export(t, once), export(t, twice)))
}

func TestRotateFullTurnIsANoOp(t *testing.T) {
	original := loadGrid(t)
	defer original.Close()
	before := export(t, original)

	require.NoError(t, RotateCommand(original, "360"))

	require.True(t, bytes.Equal(before, export(t, original)))
}

func TestRotateRejectsOutOfRange(t *testing.T) {
	image := loadGrid(t)
	defer image.Close()

	for _, args := range []string{"-1", "361", "abc"} {
		require.Error(t, RotateCommand(image, args), "rotate/%s must be refused", args)
	}
}

// A single colour image has no range to stretch. Dividing by it made the
// scale infinite and the offset not a number.
func TestAutolevelOnASingleColour(t *testing.T) {
	vips.Startup(nil)

	flat, err := vips.Black(64, 64)
	require.NoError(t, err)
	defer flat.Close()
	require.NoError(t, flat.Linear1(1, 128))

	before := export(t, flat)
	require.NoError(t, AutolevelCommand(flat, "true"))

	require.True(t, bytes.Equal(before, export(t, flat)), "a flat image must be left alone")
}
