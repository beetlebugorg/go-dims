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
)

func Handler(kernel Kernel, config Config, w http.ResponseWriter, r *http.Request) {
	slog.Info("handleDims5()",
		"imageUrl", r.URL.Query().Get("url"),
		"clientId", r.PathValue("clientId"),
		"signature", r.PathValue("signature"),
		"timestamp", r.PathValue("timestamp"),
		"commands", r.PathValue("commands"))

	// Verify signature.
	if !config.DevelopmentMode {
		if !kernel.ValidateSignature() {
			kernel.SendError(w, 500, "signature verification failed.")

			return
		}
	}

	// Download image.
	if err := kernel.FetchImage(); err != nil {
		slog.Error("downloadImage failed.", "error", err)

		kernel.SendError(w, 500, "downloadImage failed.")

		return
	}

	// Execute Imagemagick commands.
	imageType, imageBlob, err := kernel.ProcessImage()
	if err != nil {
		slog.Error("executeImagemagick failed.", "error", err)

		kernel.SendError(w, 500, "image processing failed.")

		return
	}

	// Serve the image.
	if err := kernel.SendImage(w, imageType, imageBlob); err != nil {
		slog.Error("serveImage failed.", "error", err)

		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}
