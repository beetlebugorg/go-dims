package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func SharpenCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("SharpenCommand", "args", args)

	var geometry imagick.GeometryInfo
	flags := imagick.ParseGeometry(args, &geometry)
	if (flags & imagick.SIGMAVALUE) == 0 {
		geometry.Sigma = 1.0
	}

	return mw.SharpenImage(geometry.Rho, geometry.Sigma)
}
