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

package commands

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func StripMetadataCommand(image *vips.ImageRef, args string, ops *ExportOptions) error {
	strip := args == "true"

	ops.JpegExportParams.StripMetadata = strip
	ops.PngExportParams.StripMetadata = strip
	ops.WebpExportParams.StripMetadata = strip
	ops.GifExportParams.StripMetadata = strip
	ops.TiffExportParams.StripMetadata = strip

	if strip {
		image.RemoveMetadata()
	}

	return nil
}
