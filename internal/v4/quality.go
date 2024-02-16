package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strconv"
)

func QualityCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("QualityCommand", "args", args)

	quality, err := strconv.Atoi(args)
	if err != nil {
		return err
	}

	return mw.SetImageCompressionQuality(uint(quality))
}
