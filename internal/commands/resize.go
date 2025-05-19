package commands

import (
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func ResizeCommand(image *vips.ImageRef, args string) error {
	geo, err := geometry.ParseGeometry(args)
	if err != nil {
		return NewOperationError("resize", args, err.Error())
	}
	rect := geo.ApplyMeta(image)

	xr := float64(rect.Width) / float64(image.Width())
	yr := float64(rect.Height) / float64(image.Height())

	err = image.ResizeWithVScale(xr, yr, vips.KernelLanczos3)
	if err != nil {
		return NewOperationError("resize", args, err.Error())
	}

	return nil
}
