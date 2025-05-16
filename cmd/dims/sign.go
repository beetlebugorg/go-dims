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

package main

import (
	"fmt"
	"github.com/beetlebugorg/go-dims/pkg/dims"
)

type SignCmd struct {
	ImageURL string `arg:"" name:"imageUrl" help:"Image URL to sign. For v4 urls place any value in the signature position in the URL."`
	Encrypt  bool   `help:"Encrypt the Image URL."`
}

func (cmd *SignCmd) Run() error {
	signedUrl, err := dims.SignUrl(cmd.ImageURL)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", signedUrl)

	return nil
}
