package v5

import (
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func CropCommand(request *RequestV5, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")
	image := request.vipsImage

	// Parse Geometry
	rect := imagick.RectangleInfo{
		Width:  uint(image.Width()) / 2,
		Height: uint(image.Height()) / 2,
		X:      image.OffsetX(),
		Y:      image.OffsetY(),
	}

	imagick.ParseAbsoluteGeometry(sanitizedArgs, &rect)

	if rect.X+int(rect.Width) > image.Width() {
		rect.Width = uint(image.Width() - rect.X)
	}

	if rect.Y+int(rect.Height) > image.Height() {
		rect.Height = uint(image.Height() - rect.Y)
	}

	return image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
}
