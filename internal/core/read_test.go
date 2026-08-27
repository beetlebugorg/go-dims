package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadImageBytesAcceptsUpToTheLimit(t *testing.T) {
	data, err := ReadImageBytes(strings.NewReader(strings.Repeat("a", 64)), 64)

	require.NoError(t, err)
	require.Len(t, data, 64)
}

func TestReadImageBytesRefusesOneByteOver(t *testing.T) {
	_, err := ReadImageBytes(strings.NewReader(strings.Repeat("a", 65)), 64)

	var statusError *StatusError
	require.True(t, errors.As(err, &statusError))
	require.Equal(t, 413, statusError.StatusCode)
}

func TestReadImageBytesWithoutALimit(t *testing.T) {
	data, err := ReadImageBytes(strings.NewReader(strings.Repeat("a", 1000)), 0)

	require.NoError(t, err)
	require.Len(t, data, 1000)
}

func TestTooLarge(t *testing.T) {
	require.NoError(t, TooLarge(64, 64))
	require.NoError(t, TooLarge(1<<30, 0), "a limit of zero disables the check")
	require.Error(t, TooLarge(65, 64))
}
