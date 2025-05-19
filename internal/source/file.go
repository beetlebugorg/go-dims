package source

import (
	"context"
	"fmt"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/caarlos0/env/v10"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type fileSourceBackend struct {
	baseDir string
}

func init() {
	envConfig := core.FileSource{}
	if err := env.Parse(&envConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}

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

func (f fileSourceBackend) FetchImage(imageSource string, timeout time.Duration) (*core.Image, error) {
	var path string
	if strings.HasPrefix(imageSource, "file://../") {
		path = strings.TrimPrefix(imageSource, "file://../")
	} else {
		path = imageSource
	}

	path = filepath.Clean(path)
	path = filepath.Join(f.baseDir, path)

	imageBytes, err := readFileWithTimeout(path, timeout)
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

func readFileWithTimeout(path string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := make(chan []byte)
	errCh := make(chan error)

	go func() {
		file, err := os.Open(path)
		if err != nil {
			errCh <- err
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
