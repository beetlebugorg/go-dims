package operations

import (
	"context"
	"strings"

	"github.com/beetlebugorg/go-dims/internal/dims/geometry"
	"github.com/davidbyttow/govips/v2/vips"
)

func CropCommand(ctx context.Context, args string) error {
	sanitizedArgs := strings.ReplaceAll(args, " ", "+")

	image := ctx.Value("image").(*vips.ImageRef)

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
