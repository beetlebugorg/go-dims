package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func FlipFlopCommand(image *vips.ImageRef, args string) error {
	if args == "horizontal" {
		return image.Flip(vips.DirectionHorizontal)
	} else if args == "vertical" {
		return image.Flip(vips.DirectionVertical)
	}

	return nil
}
