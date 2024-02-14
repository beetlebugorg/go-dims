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
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

/*
Sign an image URL

The signature is the first 7 characters of an md5 hash of:

 1. timestamp
 2. secret key (from the configuration)
 3. commands (as-is from the URL)
 4. image URL

md5(1234567890mysecretresize/100x100crop/100x100http://example.com/image.jpg)
*/
func Sign(timestamp string, secret string, commands string, imageUrl string) string {
	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")

	hash := md5.New()
	io.WriteString(hash, timestamp)
	io.WriteString(hash, secret)
	io.WriteString(hash, sanitizedCommands)
	io.WriteString(hash, imageUrl)

	return fmt.Sprintf("%x", hash.Sum(nil))[0:7]
}

func SignHmacSha256(timestamp string, secret string, commands string, imageUrl string) string {
	sanitizedCommands := strings.ReplaceAll(commands, " ", "+")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(sanitizedCommands))
	mac.Write([]byte(imageUrl))

	return fmt.Sprintf("%x", mac.Sum(nil))[0:24]
}
