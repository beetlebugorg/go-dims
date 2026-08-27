package dims

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/davidbyttow/govips/v2/vips"
)

// The v4 and v5 handlers must not write to a shared configuration value.
// Run with -race.
func TestHandlerDoesNotShareConfig(t *testing.T) {
	vips.Startup(nil)

	config := *core.ReadConfig()
	config.SigningKey = "test"
	handler := NewHandler(config)

	paths := []string{
		"/v5/resize/10x10/?url=http://127.0.0.1:1/a.jpg",
		"/dims4/client/sig/123/resize/10x10/?url=http://127.0.0.1:1/a.jpg",
	}

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		for _, path := range paths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, p, nil))
			}(path)
		}
	}
	wg.Wait()
}
