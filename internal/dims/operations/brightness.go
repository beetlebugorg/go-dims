// Copyright 2024 Jeremy Collins. All rights reserved.
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
	"fmt"
	"log/slog"
	"math"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func BrightnessCommand(image *vips.ImageRef, args string) error {
	slog.Debug("BrightnessCommand", "args", args)

	geo := geometry.ParseGeometry(args)
	brightness := float64(geo.Width)
	contrast := float64(geo.Height)

	slog.Debug("BrightnessCommand", "brightness", brightness, "contrast", contrast)

	coefficients := computeIMCoefficients(brightness, contrast)

	adjustedImage, err := applyPolynomial(image, coefficients)
	if err != nil {
		return err
	}

	return image.Insert(adjustedImage, 0, 0, false, nil)
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

// applyPolynomialContrast applies a polynomial tone curve like ImageMagick's PolynomialFunction.
func applyPolynomial(img *vips.ImageRef, coeffs []float64) (*vips.ImageRef, error) {
	if len(coeffs) == 0 {
		return nil, fmt.Errorf("no coefficients provided")
	}

	// 1. Convert to float and normalize to [0,1]
	if err := img.Cast(vips.BandFormatFloat); err != nil {
		return nil, err
	}

	if err := img.Linear([]float64{1.0 / 255.0}, []float64{0}); err != nil {
		return nil, err
	}

	// 2. Evaluate polynomial using Horner's method
	// Start with 0 image (effectively: result = 0)
	acc, err := vips.Black(img.Width(), img.Height())
	if err != nil {
		return nil, err
	}

	for _, coeff := range coeffs {
		// acc = acc * img
		if err := acc.Multiply(img); err != nil {
			return nil, err
		}

		// acc = acc + coeff
		if err := acc.Linear1(1.0, coeff); err != nil {
			return nil, err
		}
	}

	// 3. Rescale back to [0,255]
	if err := acc.Linear1(255.0, 0); err != nil {
		return nil, err
	}

	// 4. Cast to 8-bit uchar
	if err := acc.Cast(vips.BandFormatUchar); err != nil {
		return nil, err
	}

	return acc, nil
}
