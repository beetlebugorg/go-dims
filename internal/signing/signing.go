package signing

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"net/http"
	"net/url"
	"strings"
)

func NewSigner(requestUrl string, config core.Config) (core.Signer, error) {
	u, err := url.Parse(requestUrl)
	if err != nil {
		return nil, err
	}

	httpRequest := &http.Request{
		URL: u,
	}

	// Commands can be v4 (/dims4/...) or v5 (/v5/...)
	if strings.HasPrefix(u.Path, "/dims4/") {
		path := u.Path[7:]
		parts := strings.SplitN(path, "/", 4)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid dims4 path format")
		}

		httpRequest.SetPathValue("clientId", parts[0])
		httpRequest.SetPathValue("signature", parts[1])
		httpRequest.SetPathValue("timestamp", parts[2])
		httpRequest.SetPathValue("commands", parts[3])

		v4Request, err := v4.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		return v4Request, nil
	} else if strings.HasPrefix(u.Path, "/v5/") {
		cmds := strings.TrimLeft(u.Path, "/v5/")

		httpRequest.SetPathValue("commands", cmds)

		v5Request, err := v5.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		return v5Request, nil
	}

	return nil, core.NewStatusError(400, "path must start with /dims4/ or /v5/: "+u.Path)
}
