// The MIT License
//
// Copyright (c) Simple Things LLC and contributors
// Copyright (c) 2025 Jeremy Collins (modified slightly for go-dims)
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use,
// copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following
// conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sourceImageDir = "../../../resources/"
const goldenImageDir = "../../../resources/golden/"

func runGoldenTest(
	t *testing.T,
	path string,
	execFn func(img *vips.ImageRef) error,
	validateFn func(img *vips.ImageRef),
	exportFn func(img *vips.ImageRef) ([]byte, *vips.ImageMetadata, error),
) []byte {
	// Set up no-op functions if nil
	if execFn == nil {
		execFn = func(*vips.ImageRef) error { return nil }
	}
	if validateFn == nil {
		validateFn = func(*vips.ImageRef) {}
	}
	if exportFn == nil {
		exportFn = func(img *vips.ImageRef) ([]byte, *vips.ImageMetadata, error) {
			return img.ExportNative()
		}
	}

	vips.Startup(nil)

	// Load image
	image, err := vips.NewImageFromFile(sourceImageDir + path)
	require.NoError(t, err, "failed to load image: %s", path)

	// Run transformation logic
	require.NoError(t, execFn(image), "execFn failed")

	// Export the transformed image
	buf, meta, err := exportFn(image)
	require.NoError(t, err, "exportFn failed")

	// Re-import for validation
	result, err := vips.NewImageFromBuffer(buf)
	require.NoError(t, err, "failed to parse exported image buffer")

	// Run validations
	validateFn(result)

	// Compare against golden
	assertGoldenImageMatch(t, path, buf, meta.Format)

	return buf
}

// assertGoldenImageMatches compares the generated image with a golden image.
//
// The golden image is a pregenerated, known good image. The comparison is
// a byte by byte comparison.
func assertGoldenImageMatch(t *testing.T, file string, buf []byte, format vips.ImageType) {
	// Extract base filename without extension
	extIndex := strings.LastIndex(file, ".")
	if extIndex < 0 {
		t.Fatalf("assertGoldenMatch: invalid file name: %s", file)
	}
	base := file[:extIndex]

	// Construct golden file name
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.TrimPrefix(testName, "TestImage_")
	env := getEnvironment()
	ext := format.FileExt()

	goldenPath := fmt.Sprintf("%s%s.%s-%s.golden%s", goldenImageDir, base, testName, env, ext)
	failedPath := fmt.Sprintf("%s%s.%s-%s.failed%s", goldenImageDir, base, testName, env, ext)

	// Check for existing golden file
	golden, err := os.ReadFile(goldenPath)
	if err == nil {
		if !bytes.Equal(buf, golden) {
			t.Logf("assertGoldenMatch: mismatch with golden file: %s", goldenPath)
			t.Logf("actual size=%d, expected size=%d", len(buf), len(golden))

			if err := os.WriteFile(failedPath, buf, 0666); err != nil {
				t.Fatalf("assertGoldenMatch: failed to write failed image: %v", err)
			}
			assert.Fail(t, "image mismatch", "wrote failed image to: %s", failedPath)
		}
		return
	}

	// No golden file found; write new one
	t.Logf("assertGoldenMatch: writing new golden file: %s", goldenPath)
	if err := os.WriteFile(goldenPath, buf, 0644); err != nil {
		t.Fatalf("assertGoldenMatch: failed to write golden file: %v", err)
	}
}

func getEnvironment() string {
	switch runtime.GOOS {
	case "linux":
		out, _ := exec.Command("lsb_release", "-cs").Output()
		if out == nil {
			// Fallback to /etc/os-release if lsb_release is not available
			out, err := exec.Command("sh", "-c", "source /etc/os-release && echo $ID").Output()
			if err != nil {
				return "linux-unknown_" + runtime.GOARCH
			}

			return "linux-" + strings.TrimSpace(string(out)) + "_" + runtime.GOARCH
		}

		strout := strings.TrimSuffix(string(out), "\n")
		return "linux-" + strout + "_" + runtime.GOARCH
	}
	// default to unknown assets otherwise
	return "ignore"
}
