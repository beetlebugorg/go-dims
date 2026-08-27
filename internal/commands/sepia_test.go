package commands

import (
	"bytes"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

func TestSepiaRejectsBadArguments(t *testing.T) {
	image := loadGrid(t)
	defer image.Close()

	for _, args := range []string{"-0.1", "1.1", "2", "abc", ""} {
		require.Error(t, SepiaCommand(image, args), "sepia/%s must be refused", args)
	}
}

// Zero strength must leave the image untouched, not merely close to it.
func TestSepiaZeroIsANoOp(t *testing.T) {
	image := loadGrid(t)
	defer image.Close()

	before := export(t, image)
	require.NoError(t, SepiaCommand(image, "0"))

	require.True(t, bytes.Equal(before, export(t, image)))
}

func TestSepiaChangesTheImage(t *testing.T) {
	image := loadGrid(t)
	defer image.Close()

	before := export(t, image)
	require.NoError(t, SepiaCommand(image, "1"))

	require.False(t, bytes.Equal(before, export(t, image)))
	require.Equal(t, 512, image.Width())
	require.Equal(t, 512, image.Height())
}

// A sepia image is a warm monochrome: red is the strongest channel and blue
// the weakest, whatever the source colour was.
func TestSepiaProducesAWarmMonochrome(t *testing.T) {
	vips.Startup(nil)

	for _, colour := range [][]float64{
		{200, 30, 30},   // red
		{30, 200, 30},   // green
		{30, 30, 200},   // blue
		{120, 120, 120}, // grey
	} {
		image, err := vips.Black(8, 8)
		require.NoError(t, err)
		require.NoError(t, image.BandJoinConst([]float64{0, 0}))
		require.NoError(t, image.Linear([]float64{0, 0, 0}, colour))
		require.NoError(t, image.Cast(vips.BandFormatUchar))

		require.NoError(t, SepiaCommand(image, "1"))

		pixel, err := image.GetPoint(4, 4)
		require.NoError(t, err)
		require.Len(t, pixel, 3)

		require.Greater(t, pixel[0], pixel[1], "red must exceed green for %v", colour)
		require.Greater(t, pixel[1], pixel[2], "green must exceed blue for %v", colour)

		image.Close()
	}
}

// Strength has to move the result steadily from the original toward the tone.
func TestSepiaStrengthIsMonotonic(t *testing.T) {
	vips.Startup(nil)

	blueAt := func(amount string) float64 {
		image, err := vips.Black(8, 8)
		require.NoError(t, err)
		defer image.Close()
		require.NoError(t, image.BandJoinConst([]float64{0, 0}))
		require.NoError(t, image.Linear([]float64{0, 0, 0}, []float64{30, 30, 200}))
		require.NoError(t, image.Cast(vips.BandFormatUchar))
		require.NoError(t, SepiaCommand(image, amount))

		pixel, err := image.GetPoint(4, 4)
		require.NoError(t, err)

		return pixel[2]
	}

	full, half, none := blueAt("1"), blueAt("0.5"), blueAt("0")

	require.Less(t, full, half, "full strength must move furthest from the original")
	require.Less(t, half, none)
}

// An image carrying alpha must keep it, since the matrix is sized to the
// band count and anything past the third band passes through.
func TestSepiaPreservesAlpha(t *testing.T) {
	image := loadGrid(t)
	defer image.Close()

	require.NoError(t, image.AddAlpha())
	require.True(t, image.HasAlpha())

	require.NoError(t, SepiaCommand(image, "0.8"))

	require.True(t, image.HasAlpha(), "alpha must survive the recombination")
	require.Equal(t, 4, image.Bands())
}
