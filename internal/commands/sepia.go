package commands

import (
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

// sepiaMatrix is the conventional sepia recombination matrix: each output
// channel is a fixed mix of the input channels, weighted so the result is a
// warm brown monochrome.
//
// This is go-dims' own rendering of the effect. mod_dims passed its argument
// to ImageMagick, whose result this does not reproduce.
var sepiaMatrix = [3][3]float64{
	{0.393, 0.769, 0.189},
	{0.349, 0.686, 0.168},
	{0.272, 0.534, 0.131},
}

// SepiaCommand applies a sepia tone. The argument is the strength of the
// effect between 0 and 1, where 0 leaves the image alone and 1 applies the
// full tone.
func SepiaCommand(image *vips.ImageRef, args string) error {
	amount, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return NewOperationError("sepia", args, err.Error())
	}

	if amount < 0 || amount > 1 {
		return NewOperationError("sepia", args, "sepia must be between 0 and 1")
	}

	if amount == 0 {
		return nil
	}

	// The matrix operates on red, green, and blue, so a greyscale or CMYK
	// source has to become sRGB first.
	if image.ColorSpace() != vips.InterpretationSRGB {
		if err := image.ToColorSpace(vips.InterpretationSRGB); err != nil {
			return NewOperationError("sepia", args, err.Error())
		}
	}

	// Recomb takes a 3x3 matrix and widens it itself when the image carries
	// alpha, so alpha needs no handling here.
	if err := image.Recomb(blendedSepia(amount)); err != nil {
		return NewOperationError("sepia", args, err.Error())
	}

	// Recombination works in floating point and can leave values outside the
	// 8 bit range, so cast back and let the cast clamp them.
	if err := image.Cast(vips.BandFormatUchar); err != nil {
		return NewOperationError("sepia", args, err.Error())
	}

	return nil
}

// blendedSepia mixes the identity matrix with the sepia matrix, so amount
// controls how far the image moves toward the full effect.
func blendedSepia(amount float64) [][]float64 {
	matrix := make([][]float64, 3)

	for row := 0; row < 3; row++ {
		matrix[row] = make([]float64, 3)

		for column := 0; column < 3; column++ {
			target := amount * sepiaMatrix[row][column]
			if row == column {
				target += 1 - amount
			}

			matrix[row][column] = target
		}
	}

	return matrix
}
