package core

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// countingReader records how much of a body was read. A test uses it to prove
// a declared size is refused before the read starts.
type countingReader struct {
	reader io.Reader
	read   int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.read += n

	return n, err
}

func withMaxSourceBytes(t *testing.T, limit int) {
	t.Helper()

	config := ReadConfig()
	original := config.MaxSourceBytes
	config.MaxSourceBytes = limit

	t.Cleanup(func() { config.MaxSourceBytes = original })
}

func requireTooLarge(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)

	var statusError *StatusError
	require.True(t, errors.As(err, &statusError), "want a StatusError, got %v", err)
	require.Equal(t, 400, statusError.StatusCode)
}

func TestReadSourceRefusesADeclaredSizeBeforeTheRead(t *testing.T) {
	withMaxSourceBytes(t, 10)

	body := &countingReader{reader: strings.NewReader(strings.Repeat("x", 100))}

	_, err := ReadSource(body, 100)
	requireTooLarge(t, err)
	require.Zero(t, body.read, "an oversized Content-Length costs nothing to refuse")
}

// An origin that declares no size, or declares one it then exceeds, is caught
// by the read itself.
func TestReadSourceRefusesABodyAboveTheLimit(t *testing.T) {
	withMaxSourceBytes(t, 10)

	_, err := ReadSource(strings.NewReader(strings.Repeat("x", 11)), -1)
	requireTooLarge(t, err)

	_, err = ReadSource(strings.NewReader(strings.Repeat("x", 11)), 1)
	requireTooLarge(t, err)
}

func TestReadSourceAcceptsABodyAtTheLimit(t *testing.T) {
	withMaxSourceBytes(t, 10)

	imageBytes, err := ReadSource(strings.NewReader(strings.Repeat("x", 10)), 10)
	require.NoError(t, err)
	require.Len(t, imageBytes, 10)
}

func TestReadSourceReadsEverythingWhenTheLimitIsZero(t *testing.T) {
	withMaxSourceBytes(t, 0)

	imageBytes, err := ReadSource(strings.NewReader(strings.Repeat("x", 4096)), 4096)
	require.NoError(t, err)
	require.Len(t, imageBytes, 4096)
}
