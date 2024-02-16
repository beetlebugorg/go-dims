package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strconv"
)

func SepiaCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("SepiaCommand", "args", args)

	threshold, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	return mw.SepiaToneImage(threshold * imagick.QUANTUM_RANGE)
}
