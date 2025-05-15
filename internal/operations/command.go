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
	"context"
	"github.com/beetlebugorg/go-dims/internal/core"
	"net/url"

	"github.com/davidbyttow/govips/v2/vips"
)

type Command struct {
	Name string
	Args string
}

// Context passed to commands.
type ExportOptions struct {
	vips.ImageType
	*vips.JpegExportParams
	*vips.PngExportParams
	*vips.WebpExportParams
	*vips.GifExportParams
	*vips.TiffExportParams
}

type VipsTransformOperation func(image *vips.ImageRef, args string) error
type VipsExportOperation func(image *vips.ImageRef, args string, opts *ExportOptions) error
type VipsRequestOperation func(image *vips.ImageRef, args string, data RequestOperation) error

type VipsCommand[T any] struct {
	Command
	Operation T
}

func PassThroughCommand(ctx context.Context, args string) error {
	return nil
}

type RequestOperation struct {
	URL    *url.URL    // The URL of the image being processed
	Config core.Config // The global configuration.
}
