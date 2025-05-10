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
	"log/slog"
	"net/http"
	"time"

	"github.com/beetlebugorg/go-dims/internal/dims/core"
)

func Handler(request Request, config core.Config, w http.ResponseWriter) {
	// Download image.
	start := time.Now()
	if err := request.FetchImage(); err != nil {
		slog.Error("kernel.FetchImage() failed.", "error", err)

		request.SendError(w, 500, "downloadImage failed.")

		return
	}
	slog.Info("kernel.FetchImage()", "duration", time.Since(start).Milliseconds())

	// Execute Imagemagick commands.
	start = time.Now()
	imageType, imageBlob, err := request.ProcessImage()
	if err != nil {
		slog.Error("kernel.ProcessImage() failed.", "error", err)

		request.SendError(w, 500, "image processing failed.")

		return
	}
	slog.Info("kernel.ProcessImage()", "duration", time.Since(start).Milliseconds())

	// Serve the image.
	start = time.Now()
	if err := request.SendImage(w, 200, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
	slog.Info("kernel.SendImage()", "duration", time.Since(start).Milliseconds())
}
