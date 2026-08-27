package source

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func withNetwork(t *testing.T, apply func(n *core.Network)) {
	t.Helper()

	config := core.ReadConfig()
	original := config.Network
	apply(&config.Network)

	t.Cleanup(func() { config.Network = original })
}

func withMaxSourceBytes(t *testing.T, limit int) {
	t.Helper()

	config := core.ReadConfig()
	original := config.MaxSourceBytes
	config.MaxSourceBytes = limit

	t.Cleanup(func() { config.MaxSourceBytes = original })
}

func requireRefused(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)

	var statusError *core.StatusError
	require.True(t, errors.As(err, &statusError), "want a StatusError, got %v", err)
	require.Equal(t, 400, statusError.StatusCode)
}

// The metadata endpoint is the address an SSRF is usually aimed at.
func TestFetchImageRefusesInstanceMetadata(t *testing.T) {
	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = false })

	_, err := NewHttpSourceBackend().FetchImage("http://169.254.169.254/latest/meta-data/", 3*time.Second)
	requireRefused(t, err)
}

func TestFetchImageRefusesLoopback(t *testing.T) {
	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = false })

	_, err := NewHttpSourceBackend().FetchImage("http://127.0.0.1:1/a.jpg", 3*time.Second)
	requireRefused(t, err)
}

func TestFetchImageRefusesNonHTTPScheme(t *testing.T) {
	_, err := NewHttpSourceBackend().FetchImage("file:///etc/passwd", 3*time.Second)
	requireRefused(t, err)
}

// A name that resolves to a private address is refused whatever it is called.
// This is what makes the check survive DNS rebinding.
func TestFetchImageRefusesNameResolvingToPrivate(t *testing.T) {
	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = false })

	_, err := NewHttpSourceBackend().FetchImage("http://localhost:1/a.jpg", 3*time.Second)
	requireRefused(t, err)
}

func TestFetchImageRefusesHostOutsideAllowlist(t *testing.T) {
	withNetwork(t, func(n *core.Network) { n.AllowedHosts = []string{"images.example.com"} })

	_, err := NewHttpSourceBackend().FetchImage("http://evil.example.net/a.jpg", 3*time.Second)
	requireRefused(t, err)
}

// A public URL that redirects into the metadata endpoint must not be followed.
// Private networks are allowed here so the first hop reaches the test server;
// the redirect target is refused by the allowlist at the second hop.
func TestFetchImageRefusesRedirectOffTheAllowlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) {
		n.AllowPrivateNetworks = true
		n.AllowedHosts = []string{"127.0.0.1"}
	})

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 3*time.Second)
	requireRefused(t, err)
}

func TestFetchImageCapsRedirectChain(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, server.URL+"/next", http.StatusFound)
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) {
		n.AllowPrivateNetworks = true
		n.MaxRedirects = 3
	})

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 5*time.Second)
	requireRefused(t, err)
}

// The guard can be opened deliberately for an origin inside the same network.
func TestFetchImageAllowsPrivateWhenEnabled(t *testing.T) {
	body := "not really an image"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = true })

	image, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 5*time.Second)
	require.NoError(t, err)
	require.Len(t, image.Bytes, len(body))
}

// An origin that declares a size above the limit is refused before the body
// is read.
func TestFetchImageRefusesADeclaredSizeAboveTheLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "4096")
		w.Write(make([]byte, 4096))
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = true })
	withMaxSourceBytes(t, 1024)

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 3*time.Second)
	requireRefused(t, err)
	require.Contains(t, err.Error(), "the origin declares 4096 bytes")
}

// A chunked response declares no size, so only the read can find it. The
// limit has to hold there as well.
func TestFetchImageRefusesAnUndeclaredBodyAboveTheLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Write(make([]byte, 4096))
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = true })
	withMaxSourceBytes(t, 1024)

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 3*time.Second)
	requireRefused(t, err)
	require.Contains(t, err.Error(), "above the 1024 byte limit")
	require.NotContains(t, err.Error(), "declares", "the response must carry no Content-Length")
}

func TestFetchImageAcceptsABodyUnderTheLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 512))
	}))
	defer server.Close()

	withNetwork(t, func(n *core.Network) { n.AllowPrivateNetworks = true })
	withMaxSourceBytes(t, 1024)

	image, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 3*time.Second)
	require.NoError(t, err)
	require.Len(t, image.Bytes, 512)
}
