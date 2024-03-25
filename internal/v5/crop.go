package v5

import (
	"github.com/davidbyttow/govips/v2/vips"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
	"strings"
)

func CropCommand(image *vips.ImageRef, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")

	// Parse Geometry
	rect := imagick.RectangleInfo{
		Width:  uint(image.Width()) / 2,
		Height: uint(image.Height()) / 2,
		X:      image.OffsetX(),
		Y:      image.OffsetY(),
	}

	imagick.ParseAbsoluteGeometry(sanitizedArgs, &rect)

	slog.Debug("CropCommand", "rect", rect)
	slog.Info("CropCommand", "image", image, "args", args, "sanitizedArgs", sanitizedArgs, "rect", rect)

	return image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
}
