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

package v5

import (
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"strconv"
)

func RotateCommand(request *RequestV5, args string) error {
	slog.Debug("RotateCommand", "args", args)

	degrees, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	return request.vipsImage.Similarity(1.0, degrees, &vips.ColorRGBA{}, 0, 0, 0, 0)
}
