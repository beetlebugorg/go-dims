package operations

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func FormatCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
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

	return nil
}
