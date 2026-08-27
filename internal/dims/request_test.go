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
