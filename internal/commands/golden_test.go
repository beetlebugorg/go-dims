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
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sourceImageDir = "../../resources/"
const goldenImageDir = "../../resources/golden/"

// updateGolden creates or rewrites golden files instead of comparing against them.
var updateGolden = flag.Bool("update", false, "create or rewrite golden image files")

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

	// Construct the golden file name. There is one set, because libvips
	// produces the same bytes on every architecture the project builds for.
	//
	// The bytes do change between libvips versions. The set is pinned to the
	// version in Dockerfile.builder, so run these tests on the builder image.
	// Regenerate the set with -update on that image when the pin moves.
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.TrimPrefix(testName, "TestImage_")
	ext := format.FileExt()

	goldenPath := fmt.Sprintf("%s%s.%s.golden%s", goldenImageDir, base, testName, ext)
	failedPath := fmt.Sprintf("%s%s.%s.failed%s", goldenImageDir, base, testName, ext)

	// Write the golden file when the caller asks for an update.
	if *updateGolden {
		t.Logf("assertGoldenMatch: writing golden file: %s", goldenPath)
		if err := os.WriteFile(goldenPath, buf, 0644); err != nil {
			t.Fatalf("assertGoldenMatch: failed to write golden file: %v", err)
		}
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("assertGoldenMatch: no golden file at %s. Run go test -update to create it.", goldenPath)
	}

	if !bytes.Equal(buf, golden) {
		t.Logf("assertGoldenMatch: mismatch with golden file: %s", goldenPath)
		t.Logf("actual size=%d, expected size=%d", len(buf), len(golden))

		if err := os.WriteFile(failedPath, buf, 0644); err != nil {
			t.Fatalf("assertGoldenMatch: failed to write failed image: %v", err)
		}
		assert.Fail(t, "image mismatch", "wrote failed image to: %s", failedPath)
	}
}
