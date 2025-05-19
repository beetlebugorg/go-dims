package http

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/commands"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"hash"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

type Request struct {
	dims.Request

	httpRequest  *http.Request
	httpResponse http.ResponseWriter
}

//-- Request/RequestContext Implementation

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	requestUrl := r.URL
	cmds := r.PathValue("commands")

	return &Request{
		Request:      *dims.NewRequest(requestUrl, cmds, config),
		httpRequest:  r,
		httpResponse: w,
	}, nil
}

func (r *Request) HashId() string {
	h := md5.New()
	h.Write([]byte(r.RawCommands))
	h.Write([]byte(r.ImageUrl))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Request) SendHeaders() {
	w := r.httpResponse

	cacheControl := r.CacheControl()
	if cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	}

	expires := r.Expires()
	if expires != "" {
		w.Header().Set("Expires", expires)
	}

	edgeControl := r.EdgeControl()
	if edgeControl != "" {
		w.Header().Set("Edge-Control", edgeControl)
	}

	contentDisposition := r.ContentDisposition()
	if contentDisposition != "" {
		w.Header().Set("Content-Disposition", contentDisposition)
	}

	etag := r.Etag()
	if etag != "" {
		w.Header().Set("ETag", etag)
	}

	if r.LastModified() != "" {
		w.Header().Set("Last-Modified", r.LastModified())
	}
}

func (r *Request) SendImage(status int, imageFormat string, imageBlob []byte) error {
	if imageBlob == nil {
		return fmt.Errorf("image is empty")
	}

	if status == http.StatusOK {
		r.SendHeaders()
	}

	// Set content type.
	r.httpResponse.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.ToLower(imageFormat)))

	// Set content length
	r.httpResponse.Header().Set("Content-Length", strconv.Itoa(len(imageBlob)))

	// Set status code.
	r.httpResponse.WriteHeader(status)

	// Write the image.
	_, err := r.httpResponse.Write(imageBlob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) SendError(err error) error {
	message := err.Error()

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var statusError *core.StatusError
	var operationError *commands.OperationError
	if errors.As(err, &statusError) {
		status = statusError.StatusCode
	} else if errors.As(err, &operationError) {
		status = operationError.StatusCode
	}

	errorImage, err := core.ErrorImage(r.Config().Error.Background)
	if err != nil {
		return err
	}

	// Send error headers.
	maxAge := r.Config().OriginCacheControl.Error
	if maxAge > 0 {
		r.httpResponse.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
		r.httpResponse.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
	}

	imageType, imageBlob, err := r.ProcessImage(errorImage, true)
	if err != nil {
		// If processing failed because of a bad command then return the image as-is.
		exportOptions := vips.NewJpegExportParams()
		exportOptions.Quality = 1
		imageBytes, _, _ := errorImage.ExportJpeg(exportOptions)

		return r.SendImage(status, "jpg", imageBytes)
	}

	if imageType == "" {
		imageType = "jpg"
	}

	return r.SendImage(status, imageType, imageBlob)
}

//-- Headers Interface

func (r *Request) CacheControl() string {
	maxAge := r.calculateMaxAge()
	if maxAge > 0 {
		return fmt.Sprintf("max-age=%d, public", r.calculateMaxAge())
	}

	return ""
}

func (r *Request) Etag() string {
	if r.SourceImage.Etag != "" {
		var h hash.Hash
		if r.Config().EtagAlgorithm == "md5" {
			h = md5.New()
		} else {
			h = sha256.New()
		}

		h.Write([]byte(r.HashId()))
		if r.SourceImage.Etag != "" {
			h.Write([]byte(r.SourceImage.Etag))
		}

		return fmt.Sprintf("%x", h.Sum(nil))
	}

	return ""
}

func (r *Request) LastModified() string {
	return r.SourceImage.LastModified
}

func (r *Request) Expires() string {
	maxAge := r.calculateMaxAge()
	if maxAge > 0 {
		return time.Now().Add(time.Duration(maxAge) * time.Second).UTC().Format(http.TimeFormat)
	}

	return ""
}

func (r *Request) EdgeControl() string {
	edgeControlTtl := r.Config().EdgeControl.DownstreamTtl
	if edgeControlTtl > 0 {
		return fmt.Sprintf("downstream-ttl=%d", edgeControlTtl)
	}

	return ""
}

func (r *Request) ContentDisposition() string {
	if r.SendContentDisposition {
		// Grab filename from imageUrl
		u, err := url.Parse(r.ImageUrl)
		if err != nil {
			return ""
		}

		filename := filepath.Base(u.Path)

		return fmt.Sprintf("attachment; filename=%s", filename)
	}

	return ""
}

func (r *Request) calculateMaxAge() int {
	maxAge := r.Config().OriginCacheControl.Default

	if r.Config().OriginCacheControl.UseOrigin {
		originMaxAge, err := sourceMaxAge(r.SourceImage.CacheControl)
		if err == nil {
			maxAge = originMaxAge

			// If below minimum, set to minimum.
			minCacheAge := r.Config().OriginCacheControl.Min
			if minCacheAge != 0 && maxAge <= minCacheAge {
				maxAge = minCacheAge
			}

			// If above maximum, set to maximum.
			maxCacheAge := r.Config().OriginCacheControl.Max
			if maxCacheAge != 0 && maxAge >= maxCacheAge {
				maxAge = maxCacheAge
			}
		}
	}

	return maxAge
}

func sourceMaxAge(header string) (int, error) {
	if header == "" {
		return 0, errors.New("empty header")
	}

	pattern, _ := regexp.Compile(`max-age=(\d+)`)
	match := pattern.FindStringSubmatch(header)
	if len(match) == 1 {
		sourceMaxAge, err := strconv.Atoi(match[0])
		if err != nil {
			return 0, errors.New("unable to convert to int")
		}

		return sourceMaxAge, nil
	}

	return 0, errors.New("max-age not found in header")
}
