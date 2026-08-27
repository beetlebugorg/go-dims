package commands

import (
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
)

func FormatCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
	imageType, ok := core.ImageTypes[args]
	if !ok {
		return NewOperationError("format", args, "unsupported output format")
	}

	opts.ImageType = imageType

	return nil
}
