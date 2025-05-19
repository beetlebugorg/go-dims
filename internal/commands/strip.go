package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func StripMetadataCommand(image *vips.ImageRef, args string, ops *ExportOptions) error {
	strip := args == "true"

	ops.JpegExportParams.StripMetadata = strip
	ops.PngExportParams.StripMetadata = strip
	ops.WebpExportParams.StripMetadata = strip
	ops.GifExportParams.StripMetadata = strip
	ops.TiffExportParams.StripMetadata = strip

	if strip {
		image.RemoveMetadata()
	}

	return nil
}
