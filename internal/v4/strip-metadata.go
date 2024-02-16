package v4

import (
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func StripMetadataCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("StripMetadataCommand")

	return mw.StripImage()
}
