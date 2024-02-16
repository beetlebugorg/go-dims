package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strconv"
)

func RotateCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("RotateCommand", "args", args)

	degrees, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	return mw.RotateImage(imagick.NewPixelWand(), degrees)
}
