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

	// Convert image to vips image.
	vipsImage, err := request.LoadImage(sourceImage)
	if err != nil {
		return err
	}

	// Execute Imagemagick commands.
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
