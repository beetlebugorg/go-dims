package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func AutolevelCommand(image *vips.ImageRef, args string) error {
	if args != "true" {
		return nil
	}

	statsOut, err := image.Copy()
	if err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}

	if err := statsOut.Stats(); err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}

	stat, _ := statsOut.GetPoint(0, 0)
	min := stat[0]

	stat, _ = statsOut.GetPoint(1, 0)
	max := stat[0]

	// Compute scale and offset to stretch to [0, 255]
	scale := 255.0 / (max - min)
	offset := -min * scale

	scales := make([]float64, image.Bands())
	offsets := make([]float64, image.Bands())
	for i := range scales {
		scales[i] = scale
		offsets[i] = offset
	}

	// Apply the linear stretch
	if err := image.Linear(scales, offsets); err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}

	return nil
}
