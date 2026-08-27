package source

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func withMaxSourceBytes(t *testing.T, limit int64) {
	t.Helper()

	config := core.ReadConfig()
	original := config.MaxSourceBytes
	config.MaxSourceBytes = limit

	t.Cleanup(func() { config.MaxSourceBytes = original })
}

func requireTooLarge(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)

	var statusError *core.StatusError
	require.ErrorAs(t, err, &statusError)
	require.Equal(t, 413, statusError.StatusCode)
}

// The origin declares a length, so the image is refused before the body is read.
func TestFetchImageRefusesDeclaredOversize(t *testing.T) {
	withMaxSourceBytes(t, 64)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 8192)))
	}))
	defer server.Close()

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 5*time.Second)
	requireTooLarge(t, err)
}

// A chunked response carries no Content-Length, so only the limited read can
// catch it. This is the case an origin can force.
func TestFetchImageRefusesUndeclaredOversize(t *testing.T) {
	withMaxSourceBytes(t, 64)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		for i := 0; i < 100; i++ {
			_, _ = w.Write([]byte(strings.Repeat("a", 100)))
			flusher.Flush()
		}
	}))
	defer server.Close()

	_, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 5*time.Second)
	requireTooLarge(t, err)
}

func TestFetchImageAcceptsWithinTheLimit(t *testing.T) {
	withMaxSourceBytes(t, 8192)

	body := strings.Repeat("a", 512)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	image, err := NewHttpSourceBackend().FetchImage(server.URL+"/a.jpg", 5*time.Second)
	require.NoError(t, err)
	require.Len(t, image.Bytes, len(body))
}
