package source

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpSourceBackend struct {
}

func init() {
	core.RegisterImageBackend(NewHttpSourceBackend())
}

func NewHttpSourceBackend() core.SourceBackend {
	return httpSourceBackend{}
}

func (backend httpSourceBackend) Name() string {
	return "http"
}

func (backend httpSourceBackend) CanHandle(imageSource string) bool {
	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		return true
	}

	return false
}

func (backend httpSourceBackend) FetchImage(imageUrl string, timeout time.Duration) (*core.Image, error) {
	slog.Debug("downloadImage", "url", imageUrl)

	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", fmt.Sprintf("go-dims/%s", core.Version))

	http.DefaultClient.Timeout = timeout
	image, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	imageSize := int(image.ContentLength)
	imageBytes, err := io.ReadAll(image.Body)
	if err != nil {
		return nil, err
	}

	sourceImage := core.Image{
		Status:       image.StatusCode,
		EdgeControl:  image.Header.Get("Edge-Control"),
		CacheControl: image.Header.Get("Cache-Control"),
		LastModified: image.Header.Get("Last-Modified"),
		Etag:         image.Header.Get("Etag"),
		Format:       vips.DetermineImageType(imageBytes),
		Size:         imageSize,
		Bytes:        imageBytes,
	}

	if image.StatusCode != 200 {
		return nil, &core.StatusError{
			Message:    fmt.Sprintf("failed to fetch image from %s", imageUrl),
			StatusCode: image.StatusCode,
		}
	}

	return &sourceImage, nil
}
