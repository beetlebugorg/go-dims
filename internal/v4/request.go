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

package v4

import (
	"crypto/md5"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	dims "github.com/beetlebugorg/go-dims/internal/http"
	"log/slog"
	"net/http"
)

type Request struct {
	*dims.Request

	clientId  string
	timestamp string
}

func NewRequest(r *http.Request, w http.ResponseWriter, config core.Config) (*Request, error) {
	clientId := r.PathValue("clientId")
	timestamp := r.PathValue("timestamp")

	request, err := dims.NewRequest(r, w, config)
	if err != nil {
		return nil, err
	}

	request.Signature = r.PathValue("signature")

	return &Request{
		Request: request,

		clientId:  clientId,
		timestamp: timestamp,
	}, nil
}

func (v4 *Request) HashId() string {
	h := md5.New()
	h.Write([]byte(v4.clientId))
	h.Write([]byte(v4.RawCommands))
	h.Write([]byte(v4.ImageUrl))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (v4 *Request) Validate() bool {
	return ValidateSignature(v4.Signature, v4.clientId, v4.timestamp, v4.ImageUrl, v4.SignedParams, v4.Config().SigningKey)
}

func ValidateSignature(signature, clientId, timestamp, imageUrl string, signedParams map[string]string, signingKey string) bool {
	h := md5.New()
	h.Write([]byte(clientId))
	h.Write([]byte(timestamp))
	h.Write([]byte(signingKey))
	h.Write([]byte(imageUrl))

	for _, signedParam := range signedParams {
		h.Write([]byte(signedParam))
	}

	expectedSignature := fmt.Sprintf("%x", h.Sum(nil))[0:7]
	if expectedSignature == signature {
		return true
	}

	slog.Error("verifySignature failed.",
		"expected", expectedSignature,
		"got", signature)

	return false
}
