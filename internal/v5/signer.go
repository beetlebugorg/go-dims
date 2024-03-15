package v5

import (
	"encoding/hex"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/dims"
	"net/http"
	"net/url"
	"strings"
)

func NewSigner(u *url.URL) (dims.UrlSigner, error) {
	httpRequest := http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	commands := strings.Replace(u.Path, "/v5/dims/", "", 1)
	commands = strings.Trim(commands, "/")
	httpRequest.SetPathValue("commands", commands)

	request := NewRequest(&httpRequest, dims.Config{
		EnvironmentConfig: dims.ReadConfig(),
		DevelopmentMode:   false,
		DebugMode:         false,
	})

	return request, nil
}

func (r *RequestV5) SignUrl() (string, error) {
	signature := hex.EncodeToString(r.Sign())

	u, err := url.Parse(r.Request.ImageUrl)
	if err != nil {
		return "", err
	}

	u.RawQuery = fmt.Sprintf("%s&sig=%s", u.RawQuery, signature)

	return u.String(), nil
}

func (r *RequestV5) ImageUrl() string {
	return r.Request.ImageUrl
}
