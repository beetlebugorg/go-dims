package dims

import (
	"net/url"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

// A ratio just above 2 used to select a shrink of 4, which decoded below the
// requested size and left the resize to scale back up.
func TestShrinkFactor(t *testing.T) {
	cases := map[int]int{
		0: 1, 1: 1,
		2: 2, 3: 2,
		4: 4, 5: 4, 7: 4,
		8: 8, 9: 8, 64: 8,
	}

	for ratio, want := range cases {
		require.Equal(t, want, shrinkFactor(ratio), "ratio %d", ratio)
	}
}

func requestWithFormat(t *testing.T, sourceFormat vips.ImageType, defaultFormat string, excluded []string) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/10x10?url=http://example.com/a.jpg")
	require.NoError(t, err)

	config := *core.ReadConfig()
	config.OutputFormat.Default = defaultFormat
	config.OutputFormat.Excluded = excluded

	request, err := NewRequest(u, "resize/10x10", config)
	require.NoError(t, err)
	request.SourceImage.Format = sourceFormat

	return request
}

// The list names source formats, as mod_dims reads the input format before
// deciding whether to apply its default.
func TestExcludedOutputFormats(t *testing.T) {
	// No exclusions: the default applies to everything.
	require.Equal(t, vips.ImageTypeWEBP,
		requestWithFormat(t, vips.ImageTypeGIF, "webp", nil).outputFormat())

	// A GIF source opts out, so it keeps its own handling.
	require.NotEqual(t, vips.ImageTypeWEBP,
		requestWithFormat(t, vips.ImageTypeGIF, "webp", []string{"gif"}).outputFormat())

	// A JPEG source is unaffected by a GIF exclusion.
	require.Equal(t, vips.ImageTypeWEBP,
		requestWithFormat(t, vips.ImageTypeJPEG, "webp", []string{"gif"}).outputFormat())

	// Matching ignores case and surrounding spaces.
	require.NotEqual(t, vips.ImageTypeWEBP,
		requestWithFormat(t, vips.ImageTypeGIF, "webp", []string{" GIF ", "svg"}).outputFormat())
}

// An excluded GIF still falls through to the PNG special case rather than
// being served as a GIF the encoder may not support.
func TestExcludedGifStillBecomesPng(t *testing.T) {
	require.Equal(t, vips.ImageTypePNG,
		requestWithFormat(t, vips.ImageTypeGIF, "webp", []string{"gif"}).outputFormat())
}

func imageOfSize(t *testing.T, width, height int) *vips.ImageRef {
	t.Helper()

	vips.Startup(nil)

	image, err := vips.Black(width, height)
	require.NoError(t, err)

	return image
}

// Width and height are metadata, so the check runs before any pixel is
// produced. That is what makes it a cheap refusal rather than an expensive one.
func TestCheckPixels(t *testing.T) {
	image := imageOfSize(t, 1000, 1000)
	defer image.Close()

	require.NoError(t, checkPixels(image, 1_000_000, "source"), "exactly at the limit is allowed")
	require.NoError(t, checkPixels(image, 2_000_000, "source"))
	require.NoError(t, checkPixels(image, 0, "source"), "zero disables the check")
	require.NoError(t, checkPixels(image, -1, "source"))

	err := checkPixels(image, 999_999, "output")
	require.Error(t, err)

	var statusError *core.StatusError
	require.ErrorAs(t, err, &statusError)
	require.Equal(t, 400, statusError.StatusCode)
	require.Contains(t, err.Error(), "output")
	require.Contains(t, err.Error(), "1000 by 1000")
}

// An upscale asks for far more work than its source suggests, which is the
// case a source byte limit cannot see.
func TestOutputCapRefusesAnUpscale(t *testing.T) {
	config := *core.ReadConfig()
	config.MaxOutputPixels = 1_000_000

	u, err := url.Parse("http://localhost/v5/resize/10000x10000?url=http://example.com/a.jpg")
	require.NoError(t, err)

	request, err := NewRequest(u, "resize/10000x10000", config)
	require.NoError(t, err)

	image := imageOfSize(t, 512, 512)
	defer image.Close()
	require.NoError(t, image.BandJoinConst([]float64{0, 0}))

	_, _, err = request.ProcessImage(image, false)
	require.Error(t, err, "a 100 megapixel output must be refused under a 1 megapixel cap")
	require.Contains(t, err.Error(), "output")
}

func TestOutputCapAllowsOrdinaryWork(t *testing.T) {
	config := *core.ReadConfig()
	config.MaxOutputPixels = 50_000_000

	u, err := url.Parse("http://localhost/v5/resize/200x200?url=http://example.com/a.jpg")
	require.NoError(t, err)

	request, err := NewRequest(u, "resize/200x200", config)
	require.NoError(t, err)

	image := imageOfSize(t, 2000, 2000)
	defer image.Close()
	require.NoError(t, image.BandJoinConst([]float64{0, 0}))

	_, _, err = request.ProcessImage(image, false)
	require.NoError(t, err)
}
