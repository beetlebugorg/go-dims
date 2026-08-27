package dims

import (
	"errors"
	"net/url"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func signedQueryOf(t *testing.T, rawQuery string) string {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/100x100?" + rawQuery)
	require.NoError(t, err)

	return SignedQuery(u)
}

// The parameters the signature never covers stay out of the canonical query.
func TestSignedQueryDropsTheExcludedParameters(t *testing.T) {
	got := signedQueryOf(t, "url=http://a/b.jpg&sig=abc&eurl=x&_keys=a&download=1&a=1")

	require.Equal(t, "a=1", got)
}

// A parameter carrying several values contributes each of them.
func TestSignedQueryKeepsRepeatedValues(t *testing.T) {
	require.Equal(t, "a=1&a=2&b=3", signedQueryOf(t, "b=3&a=1&a=2"))
}

// A field holding a line break could stand in for two, since the signed
// message puts one field per line.
func TestControlCharacterInASignedFieldIsRefused(t *testing.T) {
	u, err := url.Parse("http://localhost/v5/resize/100x100?url=" + url.QueryEscape("http://a/b\njpg"))
	require.NoError(t, err)

	_, err = NewRequest(u, "resize/100x100", *core.ReadConfig())

	var statusError *core.StatusError
	require.True(t, errors.As(err, &statusError), "want a StatusError, got %v", err)
	require.Equal(t, 400, statusError.StatusCode)
}

func TestControlCharacterInTheCommandPathIsRefused(t *testing.T) {
	u, err := url.Parse("http://localhost/v5/resize/100x100?url=http://a/b.jpg")
	require.NoError(t, err)

	_, err = NewRequest(u, "resize/100x100\nstrip/true", *core.ReadConfig())

	var statusError *core.StatusError
	require.True(t, errors.As(err, &statusError), "want a StatusError, got %v", err)
	require.Equal(t, 400, statusError.StatusCode)
}
