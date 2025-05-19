package commands

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func BrightnessCommand(image *vips.ImageRef, args string) error {
	geo, err := geometry.ParseGeometry(args)
	if err != nil {
		return NewOperationError("brightness", args, err.Error())
	}

	brightness := float64(geo.Width)
	contrast := float64(geo.Height)

	coefficients := computeIMCoefficients(brightness, contrast)

	adjustedImage, err := applyPolynomial(image, coefficients)
	if err != nil {
		return NewOperationError("brightness", args, err.Error())
	}

	err = image.Insert(adjustedImage, 0, 0, false, nil)
	if err != nil {
		return NewOperationError("brightness", args, err.Error())
	}

	return nil
}

// PerceptibleReciprocal prevents division by zero.
func perceptibleReciprocal(x float64) float64 {
	const epsilon = 1.0e-6
	if math.Abs(x) >= epsilon {
		return 1.0 / x
	}
	return 1.0 / epsilon
}

// Compute polynomial coefficients.
func computeIMCoefficients(brightness, contrast float64) []float64 {
	var slope float64
	if contrast < 0.0 {
		slope = 0.01*contrast + 1.0
	} else {
		slope = 100.0 * perceptibleReciprocal(100.0-contrast)
	}
	intercept := (0.01*brightness-0.5)*slope + 0.5

	return []float64{slope, intercept}
}

// applyPolynomial applies a polynomial tone curve like ImageMagick's PolynomialFunction.
func applyPolynomial(image *vips.ImageRef, coeffs []float64) (*vips.ImageRef, error) {
	if len(coeffs) == 0 {
		return nil, fmt.Errorf("no coefficients provided")
	}

	// 1. Convert to float and normalize to [0,1]
	if err := image.Cast(vips.BandFormatFloat); err != nil {
		return nil, err
	}

	if err := image.Linear1(1.0/255.0, 0); err != nil {
		return nil, err
	}

	// 2. Evaluate polynomial using Horner's method
	// Start with 0 image (effectively: result = 0)
	adjustedImage, err := vips.Black(image.Width(), image.Height())
	if err != nil {
		return nil, err
	}

	for _, coeff := range coeffs {
		// acc = acc * img
		if err := adjustedImage.Multiply(image); err != nil {
			return nil, err
		}

		// acc = acc + coeff
		if err := adjustedImage.Linear1(1.0, coeff); err != nil {
			return nil, err
		}
	}

	// 3. Rescale back to [0,255]
	if err := adjustedImage.Linear1(255.0, 0); err != nil {
		return nil, err
	}

	// 4. Cast to 8-bit uchar
	if err := adjustedImage.Cast(vips.BandFormatUchar); err != nil {
		return nil, err
	}

	return adjustedImage, nil
}
