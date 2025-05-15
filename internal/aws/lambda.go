// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	dims.RequestContext
}

func NewRequest(event events.LambdaFunctionURLRequest, config core.Config) (*Request, error) {
	requestUrl, err := url.Parse(event.RawPath + "?" + event.RawQueryString)
	if err != nil {
		return nil, err
	}

	request := &Request{}
	httpRequest := &http.Request{
		URL: requestUrl,
	}

	// Commands can be v4 (/dims4/...) or v5 (/v5/...)
	if strings.HasPrefix(requestUrl.Path, "/dims4/") {
		path := strings.TrimLeft(requestUrl.Path, "/dims4/")
		parts := strings.SplitN(path, "/", 4)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid dims4 path format")
		}

		cmds := strings.Join(parts[4:], "/")

		httpRequest.SetPathValue("clientId", parts[1])
		httpRequest.SetPathValue("signature", parts[2])
		httpRequest.SetPathValue("timestamp", parts[3])
		httpRequest.SetPathValue("commands", cmds)

		v4Request, err := v4.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		request.RequestContext = v4Request
	} else if strings.HasPrefix(requestUrl.Path, "/v5/") {
		cmds := strings.TrimLeft(requestUrl.Path, "/v5/")

		httpRequest.SetPathValue("commands", cmds)

		v5Request, err := v5.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		request.RequestContext = v5Request
	} else {
		return nil, core.NewStatusError(400, "path must start with /dims4/ or /v5/")
	}

	return request, nil
}

func (r *Request) SendImage(status int, imageFormat string, imageBlob []byte) error {
	return nil
}
