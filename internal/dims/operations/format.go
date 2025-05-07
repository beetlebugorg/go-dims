package operations

import (
	"github.com/davidbyttow/govips/v2/vips"
	"golang.org/x/exp/slog"
)

func FormatCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
	slog.Debug("FormatCommand", "args", args)

	switch args {
	case "jpeg", "jpg":
		opts.ImageType = vips.ImageTypeJPEG
	case "png":
		opts.ImageType = vips.ImageTypePNG
	case "webp":
		opts.ImageType = vips.ImageTypeWEBP
	case "gif":
		opts.ImageType = vips.ImageTypeGIF
	case "tiff", "tif":
		opts.ImageType = vips.ImageTypeTIFF
	}

	slog.Debug("FormatCommand", "imagetype", opts.ImageType)

	return nil
}
