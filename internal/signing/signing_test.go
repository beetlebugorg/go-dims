package signing

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"github.com/stretchr/testify/require"
)

func testConfig() core.Config {
	config := *core.ReadConfig()
	config.SigningKey = "s3cret"

	return config
}

// Signing a v5 URL and verifying the result must round trip.
func TestV5SignedUrlValidates(t *testing.T) {
	config := testConfig()

	signer, err := NewSigner("http://localhost/v5/resize/100x100/?url=http://example.com/a.jpg&overlay=http://cdn.example.com/l.png", config)
	require.NoError(t, err)

	u, err := url.Parse(signer.SignedUrl())
	require.NoError(t, err)
	require.NotEmpty(t, u.Query().Get("sig"))

	r := &http.Request{URL: u}
	r.SetPathValue("commands", "resize/100x100/")

	verified, err := v5.NewRequest(r, nil, config)
	require.NoError(t, err)
	require.True(t, verified.Validate())
}

// Signing a v4 URL must put the signature in the signature position, whatever
// placeholder the caller supplied.
func TestV4SignedUrlValidates(t *testing.T) {
	config := testConfig()

	signer, err := NewSigner("http://localhost/dims4/CLIENT/placeholder/1234567890/resize/100x100?url=http://example.com/a.jpg", config)
	require.NoError(t, err)

	u, err := url.Parse(signer.SignedUrl())
	require.NoError(t, err)
	require.NotContains(t, u.Path, "placeholder")

	parts := splitPath(t, u.Path)
	r := &http.Request{URL: u}
	r.SetPathValue("clientId", parts[0])
	r.SetPathValue("signature", parts[1])
	r.SetPathValue("timestamp", parts[2])
	r.SetPathValue("commands", parts[3])

	verified, err := v4.NewRequest(r, nil, config)
	require.NoError(t, err)
	require.True(t, verified.Validate())
}

// A placeholder equal to the client id must not confuse the rewrite.
func TestV4SignedUrlWithAConfusingPlaceholder(t *testing.T) {
	config := testConfig()

	signer, err := NewSigner("http://localhost/dims4/CLIENT/CLIENT/CLIENT/resize/100x100?url=http://example.com/a.jpg", config)
	require.NoError(t, err)

	u, err := url.Parse(signer.SignedUrl())
	require.NoError(t, err)

	parts := splitPath(t, u.Path)
	require.Equal(t, "CLIENT", parts[0], "client id must stay put")
	require.NotEqual(t, "CLIENT", parts[1], "the signature must replace the placeholder")
	require.Equal(t, "CLIENT", parts[2], "timestamp must stay put")
}

// Signing must not alter the request it was handed.
func TestSignedUrlIsStable(t *testing.T) {
	config := testConfig()

	for _, raw := range []string{
		"http://localhost/v5/resize/100x100/?url=http://example.com/a.jpg",
		"http://localhost/dims4/CLIENT/placeholder/1234567890/resize/100x100?url=http://example.com/a.jpg",
	} {
		signer, err := NewSigner(raw, config)
		require.NoError(t, err)

		first := signer.SignedUrl()
		second := signer.SignedUrl()

		require.Equal(t, first, second, "signing twice must give the same URL for %s", raw)
	}
}

func splitPath(t *testing.T, path string) []string {
	t.Helper()

	parts := make([]string, 0, 4)
	rest := path[len("/dims4/"):]
	for i := 0; i < 3; i++ {
		idx := indexOf(rest, '/')
		require.GreaterOrEqual(t, idx, 0, "path %q is too short", path)
		parts = append(parts, rest[:idx])
		rest = rest[idx+1:]
	}

	return append(parts, rest)
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}

	return -1
}
