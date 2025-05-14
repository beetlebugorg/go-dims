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

package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Serve   ServeCmd      `cmd:"" help:"Runs the DIMS service."`
	Encrypt EncryptionCmd `cmd:"" help:"Encrypt an eurl."`
	Decrypt DecryptionCmd `cmd:"" help:"Decrypt an eurl."`
	Health  HealthCmd     `cmd:"" help:"Check the health of the DIMS service."`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	if err != nil {
		os.Exit(1)
	}
}
