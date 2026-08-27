package commands

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

func exportOptions() *ExportOptions {
	return &ExportOptions{
		JpegExportParams: vips.NewJpegExportParams(),
		PngExportParams:  vips.NewPngExportParams(),
		WebpExportParams: vips.NewWebpExportParams(),
		GifExportParams:  vips.NewGifExportParams(),
		TiffExportParams: vips.NewTiffExportParams(),
	}
}

// An unsupported format used to fall through to ExportNative and answer with
// Content-Type: image/unknown.
func TestFormatRejectsUnsupported(t *testing.T) {
	for _, args := range []string{"avif", "JPG", "jpeg ", "", "bmp"} {
		require.Error(t, FormatCommand(nil, args, exportOptions()), "format/%s must be refused", args)
	}
}

func TestFormatAcceptsSupported(t *testing.T) {
	for _, args := range []string{"jpg", "jpeg", "png", "webp"} {
		require.NoError(t, FormatCommand(nil, args, exportOptions()), "format/%s must be accepted", args)
	}
}

func TestQualityRange(t *testing.T) {
	for _, args := range []string{"-1", "101", "999", "abc"} {
		require.Error(t, QualityCommand(nil, args, exportOptions()), "quality/%s must be refused", args)
	}

	for _, args := range []string{"0", "50", "100"} {
		require.NoError(t, QualityCommand(nil, args, exportOptions()), "quality/%s must be accepted", args)
	}
}

func TestKnown(t *testing.T) {
	for _, name := range []string{"resize", "crop", "thumbnail", "strip", "format", "quality", "watermark"} {
		require.True(t, Known(name), "%s must be known", name)
	}

	for _, name := range []string{"reisze", "", "sepia2", "unknown"} {
		require.False(t, Known(name), "%s must not be known", name)
	}
}

// A malformed argument is a caller error, so it must carry a 400 rather than
// surfacing as a 500.
func TestWatermarkArgumentErrorsAre400(t *testing.T) {
	for _, args := range []string{"1", "2,0.5,se", "0.5,2,se", "0.5,0.5,zz", "a,b,c"} {
		_, _, _, err := parseWatermarkArgs(args)
		require.Error(t, err, "watermark/%s must be refused", args)
	}

	opacity, size, gravity, err := parseWatermarkArgs("1,.35,se")
	require.NoError(t, err)
	require.Equal(t, 1.0, opacity)
	require.Equal(t, 0.35, size)
	require.Equal(t, vips.GravitySouthEast, gravity)
}

// The invalid gravity message reported the enum default rather than what the
// caller typed.
func TestInvalidGravityNamesTheInput(t *testing.T) {
	_, _, _, err := parseWatermarkArgs("0.5,0.5,zz")

	require.Error(t, err)
	require.Contains(t, err.Error(), "zz")
}
