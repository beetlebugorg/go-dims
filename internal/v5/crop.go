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
		Width:  uint(image.Width()),
		Height: uint(image.Height()),
		X:      image.OffsetX(),
		Y:      image.OffsetY(),
	}

	imagick.ParseAbsoluteGeometry(sanitizedArgs, &rect)

	slog.Debug("CropCommand", "rect", rect)

	return image.Crop(rect.X, rect.Y, int(rect.Width), int(rect.Height))
}
