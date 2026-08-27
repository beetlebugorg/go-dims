package dims

import (
	"github.com/beetlebugorg/go-dims/internal/core"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

type Headers interface {
	Etag() string
	LastModified() string
	Expires() string
	CacheControl() string
	EdgeControl() string
	ContentDisposition() string
}

type RequestContext interface {
	Headers
	Config() core.Config
	Validate() bool
	FetchImage(timeout time.Duration) (*core.Image, error)
	NotModified() bool
	SendNotModified() error
	LoadImage(image *core.Image) (*vips.ImageRef, error)
	ProcessImage(img *vips.ImageRef, strip bool) (string, []byte, error)
	SendImage(status int, imageFormat string, imageBlob []byte) error
}

func Handler(request RequestContext) error {
	// Validate the request.
	if !request.Config().DevelopmentMode && !request.Validate() {
		return core.NewStatusError(403, "Invalid signature")
	}

	// Download image.
	timeout := time.Duration(request.Config().Timeout.Download) * time.Millisecond
	sourceImage, err := request.FetchImage(timeout)
	if err != nil {
		return err
	}

	// The etag covers the commands, the image URL, and the origin etag, so it
	// can only be computed once the source has been fetched. Answering here
	// still skips the decode, the transformation, and the encode, which is
	// where the time goes.
	if request.NotModified() {
		return request.SendNotModified()
	}

	// Take a processing slot. The download above is waiting on an origin and
	// does not need one; everything below this competes for CPU.
	release, err := defaultLimiter().acquire()
	if err != nil {
		return err
	}
	defer release()

	// Convert image to vips image.
	vipsImage, err := request.LoadImage(sourceImage)
	if err != nil {
		return err
	}
	defer vipsImage.Close()

	// Execute the commands.
	imageType, imageBlob, err := request.ProcessImage(vipsImage, false)
	if err != nil {
		return err
	}

	// Serve the image.
	if err := request.SendImage(200, imageType, imageBlob); err != nil {
		return err
	}

	return nil
}
