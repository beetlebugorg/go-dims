package v4

import (
	"errors"
	"github.com/sagikazarmark/slog-shim"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func GravityCommand(mw *imagick.MagickWand, args string) error {
	slog.Debug("GravityCommand", "args", args)

	gravityMap := map[string]imagick.GravityType{
		"northwest": imagick.GRAVITY_NORTH_WEST,
		"north":     imagick.GRAVITY_NORTH,
		"northeast": imagick.GRAVITY_NORTH_EAST,
		"west":      imagick.GRAVITY_WEST,
		"center":    imagick.GRAVITY_CENTER,
		"east":      imagick.GRAVITY_EAST,
		"southwest": imagick.GRAVITY_SOUTH_WEST,
		"south":     imagick.GRAVITY_SOUTH,
		"southeast": imagick.GRAVITY_SOUTH_EAST,
	}

	gravity, ok := gravityMap[strings.ToLower(args)]
	if !ok {
		return errors.New("unknown gravity")
	}

	return mw.SetImageGravity(gravity)
}
