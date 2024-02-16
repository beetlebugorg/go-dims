package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func BrightnessCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("BrightnessCommand", "args", args)

	var geometry imagick.GeometryInfo
	imagick.ParseGeometry(args, &geometry)

	return mw.BrightnessContrastImage(geometry.Rho, geometry.Sigma)
}
