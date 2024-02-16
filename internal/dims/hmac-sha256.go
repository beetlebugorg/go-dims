// Copyright 2024 Jeremy Collins. All rights reserved.
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

package dims

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
)

type HmacSha256Algorithm struct {
	Key string
}

// NewHmacSha256 returns a new HmacSha256Algorithm.
func NewHmacSha256(signingKey string) SignatureAlgorithm {
	return HmacSha256Algorithm{
		Key: signingKey,
	}
}

// Sign returns a signed string using the HMAC-SHA256 algorithm.
func (h HmacSha256Algorithm) Sign(commands []Command, imageUrl string) string {
	mac := hmac.New(sha256.New, []byte(h.Key))
	for _, command := range commands {
		santizedArgs := strings.ReplaceAll(command.Args, " ", "+")

		mac.Write([]byte(command.Name))
		mac.Write([]byte(santizedArgs))
	}
	mac.Write([]byte(imageUrl))

	return fmt.Sprintf("%x", mac.Sum(nil))[0:24]
}