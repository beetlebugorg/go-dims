package v4

import (
	"errors"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func NewSigner(u *url.URL) (dims.UrlSigner, error) {
	path := strings.Replace(u.Path, "/v4/dims/", "", 1)
	path = strings.Replace(path, "/dims4/", "", 1)

	pattern, _ := regexp.Compile(`([^/]+)/(\w{7})/([^/]+)/(.+)`)
	match := pattern.FindStringSubmatch(path)
	if len(match) != 5 {
		slog.Error("signV4", "error", "failed to match path", "match", match, "path", u.Path)
		return nil, errors.New("failed to match path")
	}

	httpRequest := http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	httpRequest.SetPathValue("clientId", match[1])
	httpRequest.SetPathValue("signature", match[2])
	httpRequest.SetPathValue("timestamp", match[3])
	httpRequest.SetPathValue("commands", match[4])

	var request = NewRequest(&httpRequest, dims.Config{
		EnvironmentConfig: dims.ReadConfig(),
		DevelopmentMode:   false,
		DebugMode:         false,
	})

	return request, nil
}

func (r *RequestV4) SignUrl() (string, error) {
	signature := r.Sign()

	u, err := url.Parse(r.Request.ImageUrl)
	if err != nil {
		return "", err
	}

	u.RawQuery = fmt.Sprintf("%s&sig=%s", u.RawQuery, signature)

	return u.String(), nil
}

func (r *RequestV4) ImageUrl() string {
	return r.Request.ImageUrl
}
