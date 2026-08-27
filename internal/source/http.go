package source

import (
	"context"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpSourceBackend struct {
}

// httpClient is shared by every request. The download deadline comes from the
// request context, never from a field on the client.
var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   true,
	},
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", fmt.Sprintf("go-dims/%s", core.Version))

	image, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer image.Body.Close()

	maxBytes := core.ReadConfig().MaxSourceBytes
	if err := core.TooLarge(image.ContentLength, maxBytes); err != nil {
		return nil, err
	}

	imageSize := int(image.ContentLength)
	imageBytes, err := core.ReadImageBytes(image.Body, maxBytes)
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
