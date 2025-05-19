package commands

import (
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

func QualityCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
	quality, err := strconv.Atoi(args)
	if err != nil {
		return NewOperationError("quality", args, err.Error())
	}

	opts.JpegExportParams.Quality = quality
	opts.PngExportParams.Quality = quality
	opts.WebpExportParams.Quality = quality
	opts.TiffExportParams.Quality = quality
	opts.GifExportParams.Quality = quality

	return nil
}
