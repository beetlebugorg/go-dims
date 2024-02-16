package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func AutolevelCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("AutolevelCommand", "args", args)

	if args == "true" {
		return mw.AutoLevelImage()
	}

	return nil
}
