package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func InvertCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("InvertCommand", "args", args)

	if args == "true" {
		return mw.NegateImage(false)
	}

	return nil
}
