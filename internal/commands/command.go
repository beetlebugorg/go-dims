package commands

import (
	"context"
	"github.com/beetlebugorg/go-dims/internal/core"
	"net/url"

	"github.com/davidbyttow/govips/v2/vips"
)

type Command struct {
	Name string
	Args string
}

// Context passed to commands.
type ExportOptions struct {
	vips.ImageType
	*vips.JpegExportParams
	*vips.PngExportParams
	*vips.WebpExportParams
	*vips.GifExportParams
	*vips.TiffExportParams
}

type VipsTransformOperation func(image *vips.ImageRef, args string) error
type VipsExportOperation func(image *vips.ImageRef, args string, opts *ExportOptions) error
type VipsRequestOperation func(image *vips.ImageRef, args string, data RequestOperation) error

type VipsCommand[T any] struct {
	Command
	Operation T
}

func PassThroughCommand(ctx context.Context, args string) error {
	return nil
}

type RequestOperation struct {
	URL    *url.URL    // The URL of the image being processed
	Config core.Config // The global configuration.
}

var VipsTransformCommands = map[string]VipsTransformOperation{
	"crop":             CropCommand,
	"resize":           ResizeCommand,
	"sharpen":          SharpenCommand,
	"brightness":       BrightnessCommand,
	"flipflop":         FlipFlopCommand,
	"sepia":            SepiaCommand,
	"grayscale":        GrayscaleCommand,
	"autolevel":        AutolevelCommand,
	"invert":           InvertCommand,
	"rotate":           RotateCommand,
	"thumbnail":        ThumbnailCommand,
	"legacy_thumbnail": LegacyThumbnailCommand,
}

var VipsExportCommands = map[string]VipsExportOperation{
	"strip":   StripMetadataCommand,
	"format":  FormatCommand,
	"quality": QualityCommand,
}

var VipsRequestCommands = map[string]VipsRequestOperation{
	"watermark": Watermark,
}
