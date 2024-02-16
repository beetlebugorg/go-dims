package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func GrayscaleCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("GrayscaleCommand", "args", args)

	if args == "true" {
		return mw.SetImageColorspace(imagick.COLORSPACE_GRAY)
	}

	return nil
}
