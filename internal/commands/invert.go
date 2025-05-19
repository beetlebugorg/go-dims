package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func InvertCommand(image *vips.ImageRef, args string) error {
	return image.Invert()
}
