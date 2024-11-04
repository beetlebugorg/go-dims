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

import "net/http"

type Kernel interface {
	ValidateSignature() bool
	FetchImage() error
	ProcessImage() (string, []byte, error)
	ProcessCommand(command Command) error
	SendHeaders(w http.ResponseWriter)
	SendImage(w http.ResponseWriter, status int, imageType string, imageBlob []byte) error
	SendError(w http.ResponseWriter, status int, message string)
}

type ErrorImageGenerator interface {
	GenerateErrorImage(w http.ResponseWriter, status int, message string)
}

type ImageProcessor interface {
	ProcessImage() (string, []byte, error)
}

type Commands interface {
	Commands(cmds string) []Command
}
