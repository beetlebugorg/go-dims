package commands

import (
	"github.com/beetlebugorg/go-dims/internal/geometry"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

func CropCommand(image *vips.ImageRef, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+") + "!"

	rect, err := geometry.ParseGeometry(sanitizedArgs)
	if err != nil {
		return NewOperationError("crop", args, err.Error())
	}
	rect = rect.ApplyMeta(image)

	height := rect.Y + int(rect.Height)
	if height > image.Height() {
		rect.Height = float64(image.Height()) - float64(rect.Y)
	}

	width := rect.X + int(rect.Width)
	if width > image.Width() {
		rect.Width = float64(image.Width()) - float64(rect.X)
	}

	if rect.Width <= 0 {
		return NewOperationError("crop", args, "width must be greater than 0")
	}

	if rect.Height <= 0 {
		return NewOperationError("crop", args, "height must be greater than 0")
	}

	err = image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
	if err != nil {
		return NewOperationError("crop", args, err.Error())
	}

	return nil
}
