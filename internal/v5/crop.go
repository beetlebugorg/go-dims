package v5

import (
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
)

func CropCommand(request *RequestV5, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")
	image := request.vipsImage

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
