package v4

import (
	"errors"
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func LegacyThumbnailCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("LegacyThumbnailCommand", "args", args)

	// Remove any symbols and add a trailing '^' to the geometry. This ensures
	// that the image will be at least as large as requested.
	resizedArgs := strings.TrimRight(args, "^!<>") + "^"

	// Parse Geometry
	var rect imagick.RectangleInfo

	imagick.SetGeometry(mw.Image(), &rect)
	flags := imagick.ParseMetaGeometry(resizedArgs, &rect.X, &rect.Y, &rect.Width, &rect.Height)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (resize) geometry failed")
	}

	format := mw.GetImageFormat()
	if format == "JPG" {
		factors := []float64{2.0, 1.0, 1.0}
		mw.SetSamplingFactors(factors)
	}

	if rect.Width < 200 && rect.Height < 200 {
		mw.ThumbnailImage(rect.Width, rect.Height)
	} else {
		mw.ScaleImage(rect.Width, rect.Height)
	}

	flags = imagick.ParseAbsoluteGeometry(args, &rect)
	if (flags & imagick.ALLVALUES) == 0 {
		return errors.New("parsing thumbnail (crop) geometry failed")
	}

	width := mw.GetImageWidth()
	height := mw.GetImageHeight()
	x := (width / 2) - (rect.Width / 2)
	y := (height / 2) - (rect.Height / 2)

	mw.CropImage(rect.Width, rect.Height, int(x), int(y))
	mw.SetImagePage(rect.Width, rect.Height, int(x), int(y))

	return nil
}
