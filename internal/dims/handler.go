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
)

func Handler(kernel Kernel, config Config, w http.ResponseWriter) {
	// Verify signature.
	if !config.DevelopmentMode {
		if kernel.ValidateSignature() == false {
			kernel.SendError(w, 403, "verification failed.")

			slog.Info("verification failed.")

			return
		}
	}

	// Download image.
	start := time.Now()
	if err := kernel.FetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		kernel.SendError(w, 500, "downloadImage failed.")

		return
	}
	slog.Info("downloadImage", "duration", time.Since(start))

	// Execute Imagemagick commands.
	start = time.Now()
	imageType, imageBlob, err := kernel.ProcessImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		kernel.SendError(w, 500, "image processing failed.")

		return
	}
	slog.Info("executeImagemagick", "duration", time.Since(start))

	// Serve the image.
	start = time.Now()
	if err := kernel.SendImage(w, 200, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
	slog.Info("serveImage", "duration", time.Since(start))
}
