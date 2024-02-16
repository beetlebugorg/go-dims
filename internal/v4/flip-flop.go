package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func FlipFlopCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("FlipFlopCommand", "args", args)

	if args == "horizontal" {
		return mw.FlopImage()
	} else if args == "vertical" {
		return mw.FlipImage()
	}

	return nil
}
