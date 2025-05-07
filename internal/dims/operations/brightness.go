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

package operations

import (
	"log/slog"

	"github.com/davidbyttow/govips/v2/vips"
)

func BrightnessCommand(image *vips.ImageRef, args string) error {
	slog.Debug("BrightnessCommand", "args", args)

	//image := request.vipsImage

	//var geometry imagick.GeometryInfo
	//imagick.ParseGeometry(args, &geometry)

	//return image.Linear([]float64{geometry.Rho}, []float64{geometry.Sigma})
	return nil
}
