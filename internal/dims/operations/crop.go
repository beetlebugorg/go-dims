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

	height := rect.X + int(rect.Height)
	if height > image.Width() {
		rect.Height = float64(image.Height())
	}

	width := rect.Y + int(rect.Width)
	if width > image.Height() {
		rect.Width = float64(image.Width())
	}

	return image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
}
