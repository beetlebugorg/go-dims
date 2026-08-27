package source

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/beetlebugorg/go-dims/internal/core"
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
	require.Empty(t, image.LastModified)
	require.Equal(t, []byte("image bytes"), image.Bytes)
}

func TestImageFromResponseReadsPresentFields(t *testing.T) {
	vips.Startup(nil)

	modified := time.Date(2026, 8, 27, 10, 30, 0, 0, time.UTC)
	image := imageFromResponse(&s3.GetObjectOutput{
		ETag:         aws.String(`"abc123"`),
		LastModified: aws.Time(modified),
	}, []byte("image bytes"))

	require.Equal(t, `"abc123"`, image.Etag)
	require.Equal(t, "Thu, 27 Aug 2026 10:30:00 GMT", image.LastModified)
}

// DIMS_S3_PREFIX applies to a bare key, which is what arrives when s3 is the
// default backend. An s3:// URL names its own bucket and path already.
func TestS3Resolve(t *testing.T) {
	withPrefix := s3SourceBackend{Config: core.S3{Bucket: "my-bucket", Prefix: "images/2024/"}}

	bucket, key, err := withPrefix.resolve("image.jpg")
	require.NoError(t, err)
	require.Equal(t, "my-bucket", bucket)
	require.Equal(t, "images/2024/image.jpg", key)

	bucket, key, err = withPrefix.resolve("/image.jpg")
	require.NoError(t, err)
	require.Equal(t, "images/2024/image.jpg", key)

	// A full URL is taken as given.
	bucket, key, err = withPrefix.resolve("s3://other-bucket/raw/image.jpg")
	require.NoError(t, err)
	require.Equal(t, "other-bucket", bucket)
	require.Equal(t, "raw/image.jpg", key)

	// No prefix configured.
	plain := s3SourceBackend{Config: core.S3{Bucket: "my-bucket"}}
	_, key, err = plain.resolve("image.jpg")
	require.NoError(t, err)
	require.Equal(t, "image.jpg", key)
}
