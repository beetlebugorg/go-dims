package commands

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

func TestAutolevel(t *testing.T) {
	path := "grid.png"

	runGoldenTest(
		t,
		path,
		func(img *vips.ImageRef) error {
			return AutolevelCommand(img, "true")
		},
		func(img *vips.ImageRef) {
			assert.Equal(t, 512, img.Width())
			assert.Equal(t, 512, img.Height())
		},
		nil,
	)
}
