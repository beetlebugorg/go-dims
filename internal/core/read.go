package core

import (
	"fmt"
	"io"
)

// ReadImageBytes reads a source image into memory and refuses anything above
// maxBytes. It reads one byte past the limit so an oversized source is caught
// even when the origin declares no length or lies about it.
func ReadImageBytes(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(reader)
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maxBytes {
		return nil, NewStatusError(413, fmt.Sprintf("image is larger than the %d byte limit", maxBytes))
	}

	return data, nil
}

// TooLarge reports whether a declared content length exceeds maxBytes. It lets
// a backend refuse an oversized image before reading the body.
func TooLarge(contentLength int64, maxBytes int64) error {
	if maxBytes > 0 && contentLength > maxBytes {
		return NewStatusError(413, fmt.Sprintf("image is larger than the %d byte limit", maxBytes))
	}

	return nil
}
