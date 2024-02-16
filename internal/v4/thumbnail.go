package v4

import (
	"errors"
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func ThumbnailCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("ThumbnailCommand", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo
	var exception imagick.ExceptionInfo

	slog.Info("ThumbnailCommand", "resizedArgs", resizedArgs)

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	slog.Debug("ThumbnailCommand[resize]", "rect", rect)

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	mw.ThumbnailImage(rect.Width, rect.Height)

	if (flags & imagick.PERCENTVALUE) != 0 {
		flags = imagick.ParseGravityGeometry(mw.Image(), args, &rect, &exception)
		if (flags & imagick.ALLVALUES) == 0 {
			return errors.New("parsing thumbnail (crop) geometry failed")
		}

		slog.Debug("ThumbnailCommand[crop]", "rect", rect)
		mw.CropImage(rect.Width, rect.Height, rect.X, rect.Y)
		return mw.SetImagePage(rect.Width, rect.Height, rect.X, rect.Y)
	}

	return nil
}
