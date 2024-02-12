package dims

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type Request struct {
	config Config // The global configuration.

	clientId     string  // The client ID of this request.
	signature    string  // The signature of the request.
	timestamp    int32   // The timestamp of the request.
	imageUrl     string  // The image URL that is being manipulated.
	noImageUrl   string  // The URL to the image in case of failures.
	useNoImage   bool    // Whether to use the no image URL.
	commands     string  // The unparsed commands (resize, crop, etc).
	sampleFactor float64 // The sample factor for optimizing resizing.

	// The following fields are populated after the image is downloaded.
	originalImage     []byte // The downloaded image.
	originalImageSize int64  // The original image size in bytes.
	cacheControl      string // The cache headers from the downloaded image.
	edgeControl       string // The edge control headers from the downloaded image.
	lastModified      string // The last modified header from the downloaded image.
	etag              string // The etag header from the downloaded image.
}

/*
verifySignature verifies the signature of the image resize is valid.
*
* The signature is the first 6 characters of an md5 hash of:
*
*   1. timestamp
*   2. secret key (from the configuration)
*   3. commands (as-is from the URL)
*   4. image URL
*
* Example:

*   1. timestamp: 1234567890
*   2. secret
*   3. commands: resize/100x100/crop/100x100
*   4. image URL: http://example.com/image.jpg
*
*   md5(1234567890secretresize/100x100crop/100x100http://example.com/image.jpg)
*/
func (r *Request) verifySignature() error {
	slog.Info("verifySignature", "url", r.imageUrl)
	return nil
}

func (r *Request) fetchImage() error {
	slog.Info("downloadImage", "url", r.imageUrl)

	image, err := http.Get(r.imageUrl)
	if err != nil {
		return err
	}

	r.cacheControl = image.Header.Get("Cache-Control")
	r.edgeControl = image.Header.Get("Edge-Control")
	r.lastModified = image.Header.Get("Last-Modified")
	r.etag = image.Header.Get("Etag")

	r.originalImage, err = io.ReadAll(image.Body)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) processImage() (string, []byte, error) {
	slog.Debug("executeImagemagick", "commands", r.commands)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image.
	err := mw.ReadImageBlob(r.originalImage)
	if err != nil {
		return "", nil, err
	}

	// Convert image to RGB from CMYK.
	if mw.GetImageColorspace() == imagick.COLORSPACE_CMYK {

		profiles := mw.GetImageProfiles("icc")
		if profiles != nil {
			mw.ProfileImage("ICC", CmykIccProfile)
		}
		mw.ProfileImage("ICC", RgbIccProfile)

		err = mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
		if err != nil {
			return "", nil, err
		}
	}

	// Flip image orientation, if needed.
	mw.AutoOrientImage()

	// Truncate images (i.e animated gif). This removes all but the first image.
	images := mw.GetNumberImages()
	if images > 1 {
		for i := 1; i <= int(images); i++ {
			if i > 0 {
				mw.SetIteratorIndex(i)
				mw.RemoveImage()
			}
		}
	}

	// Execute the commands.
	stripMetadata := true
	explodedCommands := strings.Split(r.commands, "/")
	for i := 0; i < len(explodedCommands)-1; i += 2 {
		command := explodedCommands[i]
		args := explodedCommands[i+1]

		if command == "strip" {
			stripMetadata = false
		}

		// Lookup command, call it.
		if operation, ok := Operations[command]; ok {
			slog.Info("executeCommand", "command", command, "args", args)

			if err := operation(mw, args); err != nil {
				return "", nil, err
			}
		}
	}

	// Strip metadata. (if not already stripped)
	if stripMetadata && r.config.StripMetadata {
		mw.StripImage()
	}

	return mw.GetImageFormat(), mw.GetImageBlob(), nil
}

func (r *Request) sendImage(w http.ResponseWriter, imageType string, imageBlob []byte) error {
	// Set headers.

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", imageType))

	// Write the image.
	w.Write(imageBlob)

	return nil
}
