package dims

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type Request struct {
	config Config // The global configuration.

	clientId            string      // The client ID of this request.
	signature           string      // The signature of the request.
	timestamp           int32       // The timestamp of the request.
	imageUrl            string      // The image URL that is being manipulated.
	placeholderImageUrl string      // The URL to the image in case of failures.
	commands            string      // The unparsed commands (resize, crop, etc).
	requestHash         string      // The hash of the request.
	sampleFactor        float64     // The sample factor for optimizing resizing.
	SourceImage         SourceImage // The source image.
}

type SourceImage struct {
	originalImage       []byte // The downloaded image.
	originalImageSize   int    // The original image size in bytes.
	originalImageFormat string // The original image format.
	status              int    // The HTTP status code of the downloaded image.
	cacheControl        string // The cache headers from the downloaded image.
	edgeControl         string // The edge control headers from the downloaded image.
	lastModified        string // The last modified header from the downloaded image.
	etag                string // The etag header from the downloaded image.
	placeholder         bool   // The image is a placeholder.
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
func Sign(timestamp string, secret string, commands string, imageUrl string) string {
	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")

	md5 := md5.New()
	io.WriteString(md5, timestamp)
	io.WriteString(md5, secret)
	io.WriteString(md5, sanitizedCommands)
	io.WriteString(md5, imageUrl)

	return fmt.Sprintf("%x", md5.Sum(nil))[0:7]
}

func (r *Request) verifySignature() error {
	slog.Info("verifySignature", "url", r.imageUrl)

	hash := Sign(fmt.Sprintf("%d", r.timestamp), r.config.SecretKey, r.commands, r.imageUrl)

	if hash != r.signature {
		slog.Error("verifySignature failed.", "expected", hash, "got", r.signature)

		return fmt.Errorf("invalid signature")
	}

	return nil
}

func _fetchImage(url string) SourceImage {
	image, err := http.Get(url)
	if err != nil || image.StatusCode != 200 {
		return SourceImage{
			status: image.StatusCode,
		}
	}

	sourceImage := SourceImage{
		status:              image.StatusCode,
		edgeControl:         image.Header.Get("Edge-Control"),
		cacheControl:        image.Header.Get("Cache-Control"),
		lastModified:        image.Header.Get("Last-Modified"),
		etag:                image.Header.Get("Etag"),
		originalImageFormat: image.Header.Get("Content-Type"),
	}

	sourceImage.originalImageSize = int(image.ContentLength)
	sourceImage.originalImage, _ = io.ReadAll(image.Body)
	sourceImage.placeholder = false

	return sourceImage
}

func (r *Request) fetchImage() error {
	slog.Info("downloadImage", "url", r.imageUrl)

	r.SourceImage = _fetchImage(r.imageUrl)

	if r.SourceImage.status != 200 {
		return fmt.Errorf("failed to download image")
	}

	return nil
}

func (r *Request) processImage() (string, []byte, error) {
	slog.Debug("executeImagemagick", "commands", r.commands)

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Read the image.
	err := mw.ReadImageBlob(r.SourceImage.originalImage)
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

		// If the placeholder image is being used don't execute crop operations.
		if r.SourceImage.placeholder && (command == "crop" || command == "thumbnail") {
			var rect imagick.RectangleInfo
			imagick.ParseAbsoluteGeometry(args, &rect)

			if rect.Width > 0 && rect.Height == 0 {
				args = fmt.Sprintf("%d", rect.Height)
			} else if rect.Height > 0 && rect.Width == 0 {
				args = fmt.Sprintf("x%d", rect.Width)
			} else if rect.Width > 0 && rect.Height > 0 {
				args = fmt.Sprintf("%dx%d", rect.Width, rect.Height)
			} else {
				return "", nil, errors.New("Bad arguments")
			}

			command = "resize"
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

func (r *Request) sendPlaceholderImage(w http.ResponseWriter) {
	slog.Info("sendPlaceHolderImage", "url", r.placeholderImageUrl)

	r.SourceImage = _fetchImage(r.placeholderImageUrl)
	r.SourceImage.placeholder = true

	imageType, imageBlob, _ := r.processImage()
	r.sendImage(w, imageType, imageBlob)
}

func (r *Request) sendImage(w http.ResponseWriter, imageType string, imageBlob []byte) error {
	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	cacheControlMaxAge := r.config.CacheControlMaxAge
	edgeControlTtl := r.config.EdgeControlDownstreamTtl

	sourceMaxAge := sourceMaxAge(r.SourceImage.cacheControl)
	trustSourceImage := false

	// Do we trust the source image (i.e. we control the origin) and are we able to pull out
	// the max-age from its Cache-Control header?
	if r.config.TrustSrc && sourceMaxAge > 0 {

		// Do we have valid min and max cache control values?
		if r.config.MinSrcCacheControl >= -1 && r.config.MaxSrcCacheControl >= -1 {

			// Is the max-age value between the min and max? Use the source image value.
			if (sourceMaxAge >= r.config.MinSrcCacheControl || r.config.MinSrcCacheControl == -1) &&
				(sourceMaxAge <= r.config.MaxSrcCacheControl || r.config.MaxSrcCacheControl == -1) {
				trustSourceImage = true
			}
		}

		if trustSourceImage {
			cacheControlMaxAge = sourceMaxAge
		}
	}

	if r.SourceImage.placeholder {
		cacheControlMaxAge = r.config.PlaceholderImageExpire
	}

	// Set content type.
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", imageType))

	// Set cache headers.
	if cacheControlMaxAge > 0 {
		slog.Info("sendImage", "cacheControlMaxAge", cacheControlMaxAge)

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", cacheControlMaxAge))
		w.Header().Set("Expires",
			fmt.Sprintf("%s", time.Now().
				Add(time.Duration(cacheControlMaxAge)*time.Second).
				UTC().
				Format(http.TimeFormat)))
	}

	if edgeControlTtl > 0 {
		w.Header().Set("Edge-Control", fmt.Sprintf("downstream-ttl=%d, public", edgeControlTtl))
	}

	// Set etag header.
	if r.SourceImage.etag != "" || r.SourceImage.lastModified != "" {
		md5 := md5.New()
		io.WriteString(md5, r.requestHash)

		if r.SourceImage.etag != "" {
			io.WriteString(md5, r.SourceImage.etag)
		} else if r.SourceImage.lastModified != "" {
			io.WriteString(md5, r.SourceImage.lastModified)
		}

		etag := fmt.Sprintf("%x", md5.Sum(nil))

		w.Header().Set("Etag", etag)
	}

	// Set content length
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Write the image.
	w.Write(imageBlob)

	return nil
}

func sourceMaxAge(header string) int {
	if header == "" {
		return 0
	}

	regex, _ := regexp.Compile("max-age=([\\d]+)")
	match := regex.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0
		}

		return sourceMaxAge
	}

	return 0
}
