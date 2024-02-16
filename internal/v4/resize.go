package v4

import (
	"errors"
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func ResizeCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("ResizeCommand", "args", args)

	// Parse Geometry
	var rect imagick.RectangleInfo

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(args, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	slog.Debug("ResizeCommand", "width", rect.Width, "height", rect.Height)

	return mw.ScaleImage(rect.Width, rect.Height)
}
