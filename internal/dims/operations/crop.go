package operations

import (
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func CropCommand(image *vips.ImageRef, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")

	// Parse Geometry
	rect := geometry.ParseGeometry(sanitizedArgs)

	height := rect.X + int(rect.Width)
	if height > image.Width() {
		height = image.Width()
	}

	width := rect.Y + int(rect.Height)
	if width > image.Height() {
		width = image.Height()
	}

	return image.Crop(rect.X, rect.Y, width, height)
}
