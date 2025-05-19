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

	idx, idy, odx, ody := 0.0, 0.0, 0.0, 0.0
	if degrees == 90 {
		idx, idy = 0.0, 1.0
		odx, ody = 1.0, 0.0
	} else if degrees == 180 {
		idx, idy = 0.0, 1.0
		odx, ody = 1.0, 0.0
	} else if degrees == 270 {
		idx, idy = 1.0, 0.0
		odx, ody = 0.0, -1.0
	}

	return image.Similarity(1.0, degrees, &vips.ColorRGBA{}, idx, idy, odx, ody)
}
