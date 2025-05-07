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

	if rect.X+rect.Width > image.Width() {
		rect.Width = image.Width()
	}

	if rect.Y+rect.Height > image.Height() {
		rect.Height = image.Height()
	}

	return image.Crop(rect.X, rect.Y, rect.Width, rect.Height)
}
