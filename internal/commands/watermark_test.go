package commands

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

// reduceOpacity closes the alpha copy it builds. libvips operations are lazy,
// so the export below is what proves the pipeline still holds its own
// reference after the Go handle goes away.
func TestReduceOpacityMaterializes(t *testing.T) {
	vips.Startup(nil)

	image, err := vips.NewImageFromFile(sourceImageDir + "grid.png")
	require.NoError(t, err, "failed to load image")
	defer image.Close()

	require.NoError(t, reduceOpacity(image, 0.35))
	require.True(t, image.HasAlpha())

	buf, _, err := image.ExportNative()
	require.NoError(t, err, "export after closing the alpha copy")
	require.NotEmpty(t, buf)
}
