package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func newTestRequest(t *testing.T) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/10x10?url=http://example.com/a.jpg")
	require.NoError(t, err)

	r := &http.Request{URL: u}
	r.SetPathValue("commands", "resize/10x10")

	request, err := NewRequest(r, httptest.NewRecorder(), *core.ReadConfig())
	require.NoError(t, err)

	return request
}

// RFC 9110 requires a quoted string. Proxies drop an unquoted ETag, which
// turns off revalidation without saying so.
func TestEtagIsQuoted(t *testing.T) {
	request := newTestRequest(t)
	request.SourceImage.Etag = `"origin-etag"`

	etag := request.Etag()

	require.NotEmpty(t, etag)
	require.True(t, strings.HasPrefix(etag, `"`), "got %s", etag)
	require.True(t, strings.HasSuffix(etag, `"`), "got %s", etag)
	require.Greater(t, len(etag), 2)
}

func TestEtagEmptyWithoutOriginEtag(t *testing.T) {
	require.Empty(t, newTestRequest(t).Etag())
}

// A second send must be refused rather than writing a status line twice and
// appending another image to a part written response.
func TestSendImageOnlyCommitsOnce(t *testing.T) {
	request := newTestRequest(t)

	require.NoError(t, request.SendImage(200, "jpeg", []byte("first")))
	require.Error(t, request.SendImage(200, "jpeg", []byte("second")))
}

func TestSendErrorAfterCommitDoesNotWriteAgain(t *testing.T) {
	request := newTestRequest(t)

	require.NoError(t, request.SendImage(200, "jpeg", []byte("first")))
	require.Error(t, request.SendError(core.NewStatusError(500, "late failure")))
}

// The pattern never matched before: FindStringSubmatch returns the whole
// match plus each group, so a hit has two entries, not one.
func TestSourceMaxAge(t *testing.T) {
	cases := []struct {
		header string
		want   int
		ok     bool
	}{
		{"max-age=3600, public", 3600, true},
		{"public, max-age=120", 120, true},
		{"max-age=0", 0, true},
		{"public, s-maxage=99, max-age=42", 42, true},
		{"no-store", 0, false},
		{"", 0, false},
		{"max-age=abc", 0, false},
	}

	for _, test := range cases {
		got, err := sourceMaxAge(test.header)

		if !test.ok {
			require.Error(t, err, "%q must not parse", test.header)
			continue
		}

		require.NoError(t, err, "%q must parse", test.header)
		require.Equal(t, test.want, got, "%q", test.header)
	}
}
