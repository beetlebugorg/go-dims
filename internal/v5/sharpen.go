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
	"gopkg.in/gographics/imagick.v3/imagick"
	"log/slog"
)

func SharpenCommand(request *RequestV5, args string) error {
	slog.Debug("SharpenCommand", "args", args)

	var geometry imagick.GeometryInfo
	flags := imagick.ParseGeometry(args, &geometry)
	if (flags & imagick.SIGMAVALUE) == 0 {
		geometry.Sigma = 1.0
	}

	slog.Info("SharpenCommand", "geometry", geometry)

	return request.vipsImage.Sharpen(geometry.Sigma, geometry.Rho, geometry.Rho)
}
