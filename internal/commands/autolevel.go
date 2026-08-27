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
	defer statsOut.Close()

	if err := statsOut.Stats(); err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}

	stat, err := statsOut.GetPoint(0, 0)
	if err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}
	lowest := stat[0]

	stat, err = statsOut.GetPoint(1, 0)
	if err != nil {
		return NewOperationError("autolevel", args, err.Error())
	}
	highest := stat[0]

	// An image of one colour has nothing to stretch. Dividing by the range
	// would make the scale infinite and the offset not a number.
	if highest-lowest < 1e-6 {
		return nil
	}

	// Compute scale and offset to stretch to [0, 255]
	scale := 255.0 / (highest - lowest)
	offset := -lowest * scale

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
