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
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/beetlebugorg/go-dims/internal/commands"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	dims.RequestContext

	response *events.LambdaFunctionURLStreamingResponse
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
		return nil, core.NewStatusError(400, "path must start with /dims4/ or /v5/: "+requestUrl.Path)
	}

	return request, nil
}

func (r *Request) SendImage(status int, imageFormat string, imageBlob []byte) error {
	response := &events.LambdaFunctionURLStreamingResponse{}

	headers := make(map[string]string)
	if r.RequestContext.CacheControl() != "" {
		headers["Cache-Control"] = r.CacheControl()
	}

	if r.RequestContext.Etag() != "" {
		headers["ETag"] = r.Etag()
	}

	if r.RequestContext.Expires() != "" {
		headers["Expires"] = r.Expires()
	}

	if r.RequestContext.LastModified() != "" {
		headers["Last-Modified"] = r.LastModified()
	}

	if r.RequestContext.ContentDisposition() != "" {
		headers["Content-Disposition"] = r.ContentDisposition()
	}

	if r.RequestContext.EdgeControl() != "" {
		headers["Edge-Control"] = r.EdgeControl()
	}

	headers["Content-Type"] = fmt.Sprintf("image/%s", strings.ToLower(imageFormat))
	headers["Content-Length"] = strconv.Itoa(len(imageBlob))

	response.StatusCode = status
	response.Headers = headers
	response.Body = bytes.NewReader(imageBlob)

	r.response = response

	return nil
}

func (r *Request) SendError(err error) error {
	response := &events.LambdaFunctionURLStreamingResponse{}

	message := err.Error()

	var statusError *core.StatusError
	var operationError *commands.OperationError
	if errors.As(err, &statusError) {
		response.StatusCode = statusError.StatusCode
	} else if errors.As(err, &operationError) {
		response.StatusCode = operationError.StatusCode
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = strconv.Itoa(len(message))

	response.StatusCode = 500
	response.Headers = headers
	response.Body = bytes.NewReader([]byte(message))

	return nil
}

func (r *Request) Response() *events.LambdaFunctionURLStreamingResponse {
	return r.response
}
