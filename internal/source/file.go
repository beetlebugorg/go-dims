package source

import (
	"context"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/caarlos0/env/v10"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type fileSourceBackend struct {
	baseDir string
}

func init() {
	envConfig := core.FileSource{}
	core.RecordStartupError(env.Parse(&envConfig))

	core.RegisterImageBackend(NewFileSourceBackend(envConfig.BaseDir))
}

func NewFileSourceBackend(baseDir string) core.SourceBackend {
	return fileSourceBackend{
		baseDir: baseDir,
	}
}

func (f fileSourceBackend) Name() string {
	return "file"
}

func (f fileSourceBackend) CanHandle(imageSource string) bool {
	if strings.HasPrefix(imageSource, "file://") ||
		strings.HasPrefix(imageSource, "/") ||
		strings.HasPrefix(imageSource, "./") {
		return true
	}

	return false
}

// sourcePath turns an image source into a path relative to the base
// directory. No cleaning is attempted: os.Root refuses anything that leaves
// the root, including by way of a symlink, and doing it here as well would
// only give two places for the rule to disagree.
func sourcePath(imageSource string) string {
	path := strings.TrimPrefix(imageSource, "file://")

	return strings.TrimPrefix(path, "/")
}

func (f fileSourceBackend) FetchImage(imageSource string, timeout time.Duration) (*core.Image, error) {
	imageBytes, err := readFileWithTimeout(f.baseDir, sourcePath(imageSource), timeout)
	if err != nil {
		return nil, err
	}

	return &core.Image{
		Status: 200,
		Size:   len(imageBytes),
		Bytes:  imageBytes,
		Format: vips.DetermineImageType(imageBytes),
	}, nil
}

func readFileWithTimeout(baseDir string, name string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Buffered so the reader can finish and exit after a timeout. Unbuffered
	// channels leak the goroutine, because nothing is left to receive.
	result := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		root, err := os.OpenRoot(baseDir)
		if err != nil {
			errCh <- err
			return
		}
		defer root.Close()

		file, err := root.Open(name)
		if err != nil {
			// One answer for a missing file and for a path that leaves the
			// root, so a caller cannot use the difference to map the disk.
			slog.Debug("refused a file source", "name", name, "error", err)
			errCh <- core.NewStatusError(404, "image not found")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			errCh <- err
			return
		}
		result <- data
	}()

	select {
	case <-ctx.Done():
		return nil, core.NewStatusError(504, "Timeout reading file")
	case err := <-errCh:
		return nil, err
	case data := <-result:
		return data, nil
	}
}
