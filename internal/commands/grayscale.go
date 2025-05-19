package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func GrayscaleCommand(image *vips.ImageRef, args string) error {
	return image.ToColorSpace(vips.InterpretationBW)
}
