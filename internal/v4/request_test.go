package v4

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

const (
	testSecret    = "s3cret"
	testClientID  = "TEST"
	testTimestamp = "1234567890"
	testCommands  = "resize/100x100"
	testImageURL  = "http://example.com/a.jpg"
)

// modDimsSignature reproduces the algorithm in mod_dims, src/mod_dims.c:
//
//	signature_params = expires + secret_key + commands + image_url
//	for each token in _keys, in that order: signature_params += params[token]
//	gen_hash = md5(signature_params)
//
// go-dims keeps this available under DIMS_SIGNING_COMPAT=legacy.
func modDimsSignature(expires, secret, commands, imageURL string, values []string) string {
	h := md5.New()
	h.Write([]byte(expires))
	h.Write([]byte(secret))
	h.Write([]byte(commands))
	h.Write([]byte(imageURL))
	for _, v := range values {
		h.Write([]byte(v))
	}

	return fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

func newTestRequest(t *testing.T, rawQuery, compat string) *Request {
	t.Helper()

	u, err := url.Parse("http://localhost/dims4/TEST/sig/1234567890/resize/100x100?" + rawQuery)
	require.NoError(t, err)

	r := &http.Request{URL: u}
	r.SetPathValue("clientId", testClientID)
	r.SetPathValue("signature", "0000000")
	r.SetPathValue("timestamp", testTimestamp)
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

func signatureOf(r *Request) string {
	return r.sign(testCommands, testTimestamp, testImageURL, r.SignedQuery, testSecret)
}

// Every query parameter takes part in the signature, written as name=value
// and ordered by name.
func TestSignatureCoversEveryParameter(t *testing.T) {
	request := newTestRequest(t, query("b=2&a=1&c=3"), "")

	require.Equal(t, "a=1&b=2&c=3", request.SignedQuery)
}

// _keys no longer selects what is signed, so it cannot change the digest.
func TestKeysDoesNotChangeTheSignature(t *testing.T) {
	withKeys := newTestRequest(t, query("a=1&b=2&_keys=b,a"), "")
	without := newTestRequest(t, query("a=1&b=2"), "")

	require.Equal(t, signatureOf(without), signatureOf(withKeys))
}

// Signing the same request repeatedly must produce one answer.
func TestSignatureIsStable(t *testing.T) {
	want := signatureOf(newTestRequest(t, query("a=1&b=2&c=3&d=4"), ""))

	for i := 0; i < 200; i++ {
		got := signatureOf(newTestRequest(t, query("a=1&b=2&c=3&d=4"), ""))
		require.Equal(t, want, got, "signature changed on iteration %d", i)
	}
}

// A parameter the caller never listed is now covered. This is the hole that
// let a valid signed watermark URL be replayed with a different overlay.
func TestModifiedOverlayIsRejected(t *testing.T) {
	signed := newTestRequest(t, query("overlay="+url.QueryEscape("http://cdn.example.com/logo.png")), "")

	tampered := newTestRequest(t, query("overlay="+url.QueryEscape("http://169.254.169.254/latest/meta-data/")), "")
	tampered.Signature = signatureOf(signed)

	require.False(t, tampered.Validate(), "a modified overlay must not validate")
}

// Legacy mode reproduces the mod_dims digest exactly.
func TestLegacySignatureMatchesModDims(t *testing.T) {
	request := newTestRequest(t, query("a=1&b=2&c=3&_keys=c,a,b"), "legacy")

	// _keys lists c,a,b so mod_dims concatenates 3,1,2.
	expected := modDimsSignature(testTimestamp, testSecret, testCommands, testImageURL, []string{"3", "1", "2"})
	require.Equal(t, expected, request.signLegacy(testCommands, testTimestamp, testImageURL, request.LegacyParams, testSecret))
}

// A mod_dims signature is refused unless the operator opts in. The URL below
// carries two parameters but names only one in _keys, so the two schemes
// genuinely disagree.
func TestLegacySignatureRejectedByDefault(t *testing.T) {
	legacy := modDimsSignature(testTimestamp, testSecret, testCommands, testImageURL, []string{"1"})

	request := newTestRequest(t, query("a=1&b=2&_keys=a"), "")
	request.Signature = legacy

	require.NotEqual(t, legacy, signatureOf(request), "the test case must discriminate")
	require.False(t, request.Validate())
}

func TestLegacySignatureAcceptedInCompatMode(t *testing.T) {
	legacy := modDimsSignature(testTimestamp, testSecret, testCommands, testImageURL, []string{"1"})

	request := newTestRequest(t, query("a=1&b=2&_keys=a"), "legacy")
	request.Signature = legacy

	require.NotEqual(t, legacy, signatureOf(request), "the test case must discriminate")
	require.True(t, request.Validate())
}

// Legacy mode carries the mod_dims weakness with it. An unlisted parameter
// stays unprotected, which is why it is a migration aid and not a setting to
// leave switched on.
func TestLegacyModeLeavesUnlistedParametersOpen(t *testing.T) {
	legacy := modDimsSignature(testTimestamp, testSecret, testCommands, testImageURL, nil)

	tampered := newTestRequest(t, query("overlay="+url.QueryEscape("http://169.254.169.254/")), "legacy")
	tampered.Signature = legacy

	require.True(t, tampered.Validate(), "legacy mode does not cover unlisted parameters")
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

// A value cannot carry a separator of its own, because the canonical query is
// percent-encoded.
func TestValueCannotForgeASeparator(t *testing.T) {
	forged := newTestRequest(t, query("a="+url.QueryEscape("1&b=2")), "")
	real := newTestRequest(t, query("a=1&b=2"), "")

	require.Equal(t, "a=1%26b%3D2", forged.SignedQuery)
	require.NotEqual(t, signatureOf(forged), signatureOf(real))
}
