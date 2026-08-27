package core

import (
	"fmt"
	"io"
)

// ReadSource reads a source image and refuses one above
// DIMS_MAX_SOURCE_BYTES. The pixel caps do not cover this. They run once the
// whole body is already in memory, so only a byte limit bounds what a burst
// of downloads allocates.
//
// declared is the size the origin announced. Pass a negative number when the
// origin announced none. A declared size above the limit is refused before
// the read, so an oversized body costs nothing.
func ReadSource(body io.Reader, declared int64) ([]byte, error) {
	limit := int64(ReadConfig().MaxSourceBytes)
	if limit <= 0 {
		return io.ReadAll(body)
	}

	if declared > limit {
		return nil, NewStatusError(400, fmt.Sprintf(
			"the origin declares %d bytes, which is above the %d byte limit", declared, limit))
	}

	// The extra byte separates a body that reaches the limit from one that
	// goes past it. Both stop at the same length without it.
	imageBytes, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}

	if int64(len(imageBytes)) > limit {
		return nil, NewStatusError(400, fmt.Sprintf(
			"the source image is above the %d byte limit", limit))
	}

	return imageBytes, nil
}
