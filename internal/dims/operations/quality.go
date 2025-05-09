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
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sagikazarmark/slog-shim"
)

func QualityCommand(image *vips.ImageRef, args string, opts *ExportOptions) error {
	slog.Debug("QualityCommand", "args", args)

	quality, err := strconv.Atoi(args)
	if err != nil {
		return err
	}

	switch image.Format() {
	case vips.ImageTypeJPEG:
		opts.JpegExportParams.Quality = quality
	case vips.ImageTypePNG:
		opts.PngExportParams.Quality = quality
	case vips.ImageTypeWEBP:
		opts.WebpExportParams.Quality = quality
	case vips.ImageTypeTIFF:
		opts.TiffExportParams.Quality = quality
	case vips.ImageTypeGIF:
		opts.GifExportParams.Quality = quality
	}

	return nil
}
