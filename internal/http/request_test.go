package http

import (
	"mime"
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

	return newTestRequestWithQuery(t, "url=http://example.com/a.jpg")
}

func newTestRequestWithQuery(t *testing.T, rawQuery string) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/10x10?" + rawQuery)
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

// The documented behaviour is inline unless the caller asks to download.
func TestContentDispositionInlineByDefault(t *testing.T) {
	request := newTestRequestWithQuery(t, "url=http://example.com/photo.jpg")
	request.SendContentDisposition = true

	require.Equal(t, `inline; filename=photo.jpg`, request.ContentDisposition())
}

func TestContentDispositionAttachmentOnDownload(t *testing.T) {
	request := newTestRequestWithQuery(t, "url=http://example.com/photo.jpg&download=1")

	require.True(t, request.SendContentDisposition)
	require.Equal(t, `attachment; filename=photo.jpg`, request.ContentDisposition())
}

// A URL carrying a line break is refused before a request is built, so it
// never reaches the header at all. The signature covers the image URL, and
// the signed message puts one field per line.
func TestUrlWithALineBreakIsRefused(t *testing.T) {
	u, err := url.Parse("http://localhost/v5/resize/10x10?url=" +
		url.QueryEscape("http://example.com/a\r\nX-Injected: yes.jpg"))
	require.NoError(t, err)

	r := &http.Request{URL: u}
	r.SetPathValue("commands", "resize/10x10")

	_, err = NewRequest(r, httptest.NewRecorder(), *core.ReadConfig())
	require.Error(t, err)
}

func TestContentDispositionOffByDefault(t *testing.T) {
	require.Empty(t, newTestRequest(t).ContentDisposition())
}

// The filename comes from a caller supplied URL, so it must not be able to
// add headers of its own.
func TestContentDispositionEscapesFilename(t *testing.T) {
	hostile := []string{
		`a";b.jpg`,
		`a;filename=other.jpg`,
		`a b.jpg`,
	}

	for _, name := range hostile {
		request := newTestRequestWithQuery(t, "url="+url.QueryEscape("http://example.com/"+name))
		request.SendContentDisposition = true

		got := request.ContentDisposition()

		require.NotContains(t, got, "\r", "%s", name)
		require.NotContains(t, got, "\n", "%s", name)

		// FormatMediaType returns an empty string for a value it cannot
		// encode, and SendHeaders omits an empty header. Anything it does
		// return has to parse as one well formed value.
		if got == "" {
			continue
		}

		mediatype, params, err := mime.ParseMediaType(got)
		require.NoError(t, err, "%s produced %q", name, got)
		require.Equal(t, "inline", mediatype)
		require.Len(t, params, 1)
		require.Equal(t, name, params["filename"], "the filename must survive intact")
	}
}

func TestContentDispositionWithoutAFilename(t *testing.T) {
	request := newTestRequestWithQuery(t, "url=http://example.com/")
	request.SendContentDisposition = true

	require.Empty(t, request.ContentDisposition())
}

func requestWithCondition(t *testing.T, originEtag, ifNoneMatch string) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/10x10?url=http://example.com/a.jpg")
	require.NoError(t, err)

	r := &http.Request{URL: u, Header: make(http.Header)}
	r.SetPathValue("commands", "resize/10x10")
	if ifNoneMatch != "" {
		r.Header.Set("If-None-Match", ifNoneMatch)
	}

	request, err := NewRequest(r, httptest.NewRecorder(), *core.ReadConfig())
	require.NoError(t, err)
	request.SourceImage.Etag = originEtag

	return request
}

func TestNotModified(t *testing.T) {
	// Work out the etag this request would answer with.
	etag := requestWithCondition(t, `"origin"`, "").Etag()
	require.NotEmpty(t, etag)

	require.True(t, requestWithCondition(t, `"origin"`, etag).NotModified(),
		"an exact match must be recognised")

	require.True(t, requestWithCondition(t, `"origin"`, "W/"+etag).NotModified(),
		"a weak validator must compare equal")

	require.True(t, requestWithCondition(t, `"origin"`, `"other", `+etag).NotModified(),
		"a list must be searched")

	require.True(t, requestWithCondition(t, `"origin"`, "*").NotModified(),
		"a wildcard always matches")

	require.False(t, requestWithCondition(t, `"origin"`, `"someone-elses"`).NotModified())
	require.False(t, requestWithCondition(t, `"origin"`, "").NotModified(),
		"no condition means send the image")
	require.False(t, requestWithCondition(t, "", "*").NotModified(),
		"without an origin etag there is nothing to compare")
}

// A 304 must carry the cache headers and no body.
func TestSendNotModified(t *testing.T) {
	u, err := url.Parse("http://localhost/v5/resize/10x10?url=http://example.com/a.jpg")
	require.NoError(t, err)

	r := &http.Request{URL: u, Header: make(http.Header)}
	r.SetPathValue("commands", "resize/10x10")

	recorder := httptest.NewRecorder()
	request, err := NewRequest(r, recorder, *core.ReadConfig())
	require.NoError(t, err)
	request.SourceImage.Etag = `"origin"`

	require.NoError(t, request.SendNotModified())

	require.Equal(t, http.StatusNotModified, recorder.Code)
	require.Empty(t, recorder.Body.Bytes())
	require.NotEmpty(t, recorder.Header().Get("ETag"))
	require.Empty(t, recorder.Header().Get("Content-Length"))
}

func TestSendNotModifiedOnlyCommitsOnce(t *testing.T) {
	request := requestWithCondition(t, `"origin"`, "*")

	require.NoError(t, request.SendNotModified())
	require.Error(t, request.SendNotModified())
}
