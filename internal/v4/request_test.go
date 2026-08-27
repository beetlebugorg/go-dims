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
)

// modDimsSignature reproduces the algorithm in mod_dims, src/mod_dims.c:
//
//	signature_params = expires + secret_key + commands + image_url
//	for each token in _keys, in that order: signature_params += params[token]
//	gen_hash = md5(signature_params)
//
// The values arrive in the order _keys lists them, not sorted.
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

func newTestRequest(t *testing.T, rawQuery string) *Request {
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

	request, err := NewRequest(r, nil, config)
	require.NoError(t, err)

	return request
}

// The signature must match what mod_dims computes for the same request.
func TestSignatureMatchesModDims(t *testing.T) {
	imageURL := "http://example.com/a.jpg"
	request := newTestRequest(t, "url="+url.QueryEscape(imageURL)+"&a=1&b=2&c=3&_keys=c,a,b")

	// _keys lists c,a,b so the values concatenate as 3,1,2.
	expected := modDimsSignature(testTimestamp, testSecret, testCommands, imageURL, []string{"3", "1", "2"})
	got := request.sign(testCommands, testTimestamp, imageURL, request.SignedParams, testSecret)

	require.Equal(t, expected, got)
}

// Reordering _keys must change the signature. That is what makes the order
// part of the contract rather than an accident.
func TestSignatureFollowsKeyOrder(t *testing.T) {
	imageURL := "http://example.com/a.jpg"
	escaped := url.QueryEscape(imageURL)

	first := newTestRequest(t, "url="+escaped+"&a=1&b=2&c=3&_keys=c,a,b")
	second := newTestRequest(t, "url="+escaped+"&a=1&b=2&c=3&_keys=a,b,c")

	require.Equal(t, []string{"3", "1", "2"}, first.SignedParams)
	require.Equal(t, []string{"1", "2", "3"}, second.SignedParams)

	require.NotEqual(t,
		first.sign(testCommands, testTimestamp, imageURL, first.SignedParams, testSecret),
		second.sign(testCommands, testTimestamp, imageURL, second.SignedParams, testSecret),
	)
}

// Signing the same request repeatedly must produce the same signature.
// A map made this fail roughly half the time with three keys.
func TestSignatureIsStable(t *testing.T) {
	imageURL := "http://example.com/a.jpg"
	request := newTestRequest(t, "url="+url.QueryEscape(imageURL)+"&a=1&b=2&c=3&d=4&_keys=a,b,c,d")

	want := request.sign(testCommands, testTimestamp, imageURL, request.SignedParams, testSecret)
	for i := 0; i < 200; i++ {
		fresh := newTestRequest(t, "url="+url.QueryEscape(imageURL)+"&a=1&b=2&c=3&d=4&_keys=a,b,c,d")
		got := fresh.sign(testCommands, testTimestamp, imageURL, fresh.SignedParams, testSecret)
		require.Equal(t, want, got, "signature changed on iteration %d", i)
	}
}

// A URL with no _keys must sign exactly as mod_dims does with no extra params.
func TestSignatureWithoutKeys(t *testing.T) {
	imageURL := "http://example.com/a.jpg"
	request := newTestRequest(t, "url="+url.QueryEscape(imageURL))

	require.Empty(t, request.SignedParams)
	require.Equal(t,
		modDimsSignature(testTimestamp, testSecret, testCommands, imageURL, nil),
		request.sign(testCommands, testTimestamp, imageURL, request.SignedParams, testSecret),
	)
}
