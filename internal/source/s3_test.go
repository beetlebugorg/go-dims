package source

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/require"
)

// GetObjectOutput carries every one of these as a pointer. A response that
// omits one used to panic the request goroutine.
func TestImageFromResponseHandlesNilFields(t *testing.T) {
	vips.Startup(nil)

	image := imageFromResponse(&s3.GetObjectOutput{}, []byte("image bytes"))

	require.Equal(t, 200, image.Status)
	require.Empty(t, image.Etag)
	require.Zero(t, image.Size)
	require.Empty(t, image.LastModified)
	require.Equal(t, []byte("image bytes"), image.Bytes)
}

func TestImageFromResponseReadsPresentFields(t *testing.T) {
	vips.Startup(nil)

	modified := time.Date(2026, 8, 27, 10, 30, 0, 0, time.UTC)
	image := imageFromResponse(&s3.GetObjectOutput{
		ETag:          aws.String(`"abc123"`),
		ContentLength: aws.Int64(11),
		LastModified:  aws.Time(modified),
	}, []byte("image bytes"))

	require.Equal(t, `"abc123"`, image.Etag)
	require.Equal(t, 11, image.Size)
	require.Equal(t, "Thu, 27 Aug 2026 10:30:00 GMT", image.LastModified)
}
