package commands

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

func TestBrightness(t *testing.T) {
	path := "grid.png"
	args := "10x20"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return BrightnessCommand(img, args)
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 512, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil,
	)
}
