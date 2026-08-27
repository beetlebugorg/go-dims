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

	if quality < 0 || quality > 100 {
		return NewOperationError("quality", args, "quality must be between 0 and 100")
	}

	opts.JpegExportParams.Quality = quality
	opts.PngExportParams.Quality = quality
	opts.WebpExportParams.Quality = quality
	opts.TiffExportParams.Quality = quality
	opts.GifExportParams.Quality = quality

	return nil
}
