package source

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func fileBackend(t *testing.T) (core.SourceBackend, string) {
	t.Helper()

	base := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(base, "ok.jpg"), []byte("image bytes"), 0o600))

	return NewFileSourceBackend(base), base
}

// Every one of these resolved outside the base directory before the change.
func TestFetchImageRefusesTraversal(t *testing.T) {
	backend, _ := fileBackend(t)

	sources := []string{
		"./../../etc/passwd",
		"file://../../etc/passwd",
		"./../.././../../etc/shadow",
		"/../../etc/passwd",
		"../ok.jpg",
		"file://../ok.jpg",
	}

	for _, source := range sources {
		_, err := backend.FetchImage(source, 5*time.Second)
		require.Error(t, err, "%s must be refused", source)

		var statusError *core.StatusError
		require.True(t, errors.As(err, &statusError), "%s: want a StatusError, got %v", source, err)
		require.Equal(t, 404, statusError.StatusCode, "%s", source)
	}
}

// A symlink is the case a string prefix check cannot catch.
func TestFetchImageRefusesSymlinkEscape(t *testing.T) {
	backend, base := fileBackend(t)

	outside := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(base, "link.jpg")))

	_, err := backend.FetchImage("link.jpg", 5*time.Second)
	require.Error(t, err, "a symlink leaving the root must be refused")
}

func TestFetchImageReadsInsideTheRoot(t *testing.T) {
	backend, _ := fileBackend(t)

	for _, source := range []string{"ok.jpg", "/ok.jpg", "./ok.jpg", "file://ok.jpg"} {
		image, err := backend.FetchImage(source, 5*time.Second)
		require.NoError(t, err, "%s must be read", source)
		require.Equal(t, []byte("image bytes"), image.Bytes)
	}
}

func TestFetchImageMissingFile(t *testing.T) {
	backend, _ := fileBackend(t)

	_, err := backend.FetchImage("absent.jpg", 5*time.Second)
	require.Error(t, err)
}
