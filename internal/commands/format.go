package commands

import (
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
)

func FormatCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
	opts.ImageType = core.ImageTypes[args]

	return nil
}
