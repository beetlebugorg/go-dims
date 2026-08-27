package source

import (
	"context"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

type httpSourceBackend struct {
}

// httpClient is shared by every request. The download deadline comes from the
// request context, never from a field on the client.
var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			Control:   controlAddress,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   true,
	},
	CheckRedirect: checkRedirect,
}

// controlAddress runs after the host resolves and before the socket connects,
// once per connection. Checking here rather than on the URL means a name that
// resolves to a private address is refused whatever it is called, and a
// redirect gets the same treatment as the original request.
func controlAddress(network string, address string, _ syscall.RawConn) error {
	if core.ReadConfig().AllowPrivateNetworks {
		return nil
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return core.NewStatusError(400, "cannot read the address to connect to")
	}

	ip := net.ParseIP(host)
	if !core.IsPublicAddress(ip) {
		slog.Warn("refused a non-public address", "address", host)
		return core.NewStatusError(400, "refusing to connect to a non-public address")
	}

	return nil
}

// checkRedirect applies the scheme and host rules again at every hop, and
// caps how far a chain may run.
func checkRedirect(request *http.Request, via []*http.Request) error {
	config := core.ReadConfig()
	if len(via) >= config.MaxRedirects {
		return core.NewStatusError(400, "too many redirects")
	}

	return core.ValidateImageURL(request.URL, config.Network)
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

	parsed, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return nil, err
	}

	if err := core.ValidateImageURL(parsed, core.ReadConfig().Network); err != nil {
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
