package v5

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

const (
	testSecret   = "s3cret"
	testCommands = "resize/100x100"
	testImageURL = "http://example.com/a.jpg"
)

func newTestRequest(t *testing.T, rawQuery, compat string) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/v5/resize/100x100?" + rawQuery)
	require.NoError(t, err)

	r := &http.Request{URL: u}
	r.SetPathValue("commands", testCommands)

	config := *core.ReadConfig()
	config.SigningKey = testSecret
	config.Compat = compat

	request, err := NewRequest(r, nil, config)
	require.NoError(t, err)

	return request
}

func query(extra string) string {
	q := "url=" + url.QueryEscape(testImageURL)
	if extra != "" {
		q += "&" + extra
	}

	return q
}

// legacySignature reproduces the mod_dims style scheme: only the values named
// by _keys, and a digest cut to 31 bytes.
func legacySignature(values []string) []byte {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(testCommands))
	mac.Write([]byte(testImageURL))
	for _, v := range values {
		mac.Write([]byte(v))
	}

	return mac.Sum(nil)[0:31]
}

// The digest is the full HMAC-SHA256, as the documentation states.
func TestSignatureIsFullLength(t *testing.T) {
	request := newTestRequest(t, query(""), "")

	require.Len(t, request.sign(testImageURL, request.SignedQuery, testCommands, testSecret), 32)
}

func TestSignatureCoversEveryParameter(t *testing.T) {
	request := newTestRequest(t, query("b=2&a=1&c=3"), "")

	require.Equal(t, "a=1&b=2&c=3", request.SignedQuery)
}

// Signing a URL and verifying it must round trip.
func TestSignedUrlValidates(t *testing.T) {
	request := newTestRequest(t, query("overlay="+url.QueryEscape("http://cdn.example.com/logo.png")), "")

	u, err := url.Parse(request.SignedUrl())
	require.NoError(t, err)

	r := &http.Request{URL: u}
	r.SetPathValue("commands", testCommands)

	config := *core.ReadConfig()
	config.SigningKey = testSecret

	verified, err := NewRequest(r, nil, config)
	require.NoError(t, err)
	require.True(t, verified.Validate())
}

// Replaying a signed URL with a different overlay must fail.
func TestModifiedOverlayIsRejected(t *testing.T) {
	signed := newTestRequest(t, query("overlay="+url.QueryEscape("http://cdn.example.com/logo.png")), "")

	u, err := url.Parse(signed.SignedUrl())
	require.NoError(t, err)

	q := u.Query()
	q.Set("overlay", "http://169.254.169.254/latest/meta-data/")
	u.RawQuery = q.Encode()

	r := &http.Request{URL: u}
	r.SetPathValue("commands", testCommands)

	config := *core.ReadConfig()
	config.SigningKey = testSecret

	tampered, err := NewRequest(r, nil, config)
	require.NoError(t, err)
	require.False(t, tampered.Validate(), "a modified overlay must not validate")
}

func TestLegacySignatureRejectedByDefault(t *testing.T) {
	legacy := hex.EncodeToString(legacySignature([]string{"1"}))

	request := newTestRequest(t, query("a=1&b=2&_keys=a&sig="+legacy), "")

	require.False(t, request.Validate())
}

func TestLegacySignatureAcceptedInCompatMode(t *testing.T) {
	legacy := hex.EncodeToString(legacySignature([]string{"1"}))

	request := newTestRequest(t, query("a=1&b=2&_keys=a&sig="+legacy), "legacy")

	require.True(t, request.Validate())
}

func signatureOf(r *Request) []byte {
	return r.sign(testImageURL, r.SignedQuery, testCommands, testSecret)
}

// The parameter name is part of the signed string, so moving a character from
// one parameter to the next changes the signature. Signing the values alone
// wrote "ab" then "c" as the same bytes as "a" then "bc".
func TestAdjacentParametersAreBound(t *testing.T) {
	first := newTestRequest(t, query("a=ab&b=c"), "")
	second := newTestRequest(t, query("a=a&b=bc"), "")

	require.NotEqual(t, first.SignedQuery, second.SignedQuery)
	require.NotEqual(t, signatureOf(first), signatureOf(second))
}

// The fields sit one per line, so the command path and the image URL cannot
// run together either.
func TestFieldBoundariesAreBound(t *testing.T) {
	request := newTestRequest(t, query(""), "")

	require.NotEqual(t,
		request.sign("http://example.com/a.jpg", "", "resize/100x100", testSecret),
		request.sign("//example.com/a.jpg", "", "resize/100x100http:", testSecret))
}

// A value cannot carry a separator of its own, because the canonical query is
// percent-encoded.
func TestValueCannotForgeASeparator(t *testing.T) {
	forged := newTestRequest(t, query("a="+url.QueryEscape("1&b=2")), "")
	real := newTestRequest(t, query("a=1&b=2"), "")

	require.Equal(t, "a=1%26b%3D2", forged.SignedQuery)
	require.NotEqual(t, signatureOf(forged), signatureOf(real))
}
