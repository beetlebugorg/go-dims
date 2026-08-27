package commands

import (
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

func RotateCommand(image *vips.ImageRef, args string) error {
	degrees, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return NewOperationError("rotate", args, err.Error())
	}

	if degrees < 0 || degrees > 360 {
		return NewOperationError("rotate", args, "rotate must be between 0 and 360")
	}

	// A quarter turn is an exact remapping of pixels. Similarity resamples it
	// instead, which loses detail and costs more for no benefit.
	switch degrees {
	case 0, 360:
		return nil
	case 90:
		return image.Rotate(vips.Angle90)
	case 180:
		return image.Rotate(vips.Angle180)
	case 270:
		return image.Rotate(vips.Angle270)
	}

	if err := image.Similarity(1.0, degrees, &vips.ColorRGBA{}, 0, 0, 0, 0); err != nil {
		return NewOperationError("rotate", args, err.Error())
	}

	return nil
}
