package commands

import (
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func SharpenCommand(image *vips.ImageRef, args string) error {
	geo, err := geometry.ParseGeometry(args)
	if err != nil {
		return NewOperationError("sharpen", args, err.Error())
	}

	x1 := geo.Width
	m2 := geo.Height * 2
	if m2 == 0 {
		m2 = 2.0
	}

	err = image.Sharpen(float64(0.5), x1, m2)
	if err != nil {
		return NewOperationError("sharpen", args, err.Error())
	}

	return nil
}
