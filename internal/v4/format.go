package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func FormatCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("FormatCommand", "args", args)

	return mw.SetImageFormat(args)
}
