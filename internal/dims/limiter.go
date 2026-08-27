package dims

import (
	"runtime"
	"sync"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
)

// limiter caps how many images are processed at once. Pixel limits bound what
// one request may ask for; this bounds how many may ask at the same time.
// Without it a burst of individually reasonable requests still saturates the
// machine, and every one of them slows down together.
type limiter struct {
	slots chan struct{}
	wait  time.Duration
}

func newLimiter(max int, wait time.Duration) *limiter {
	switch {
	case max < 0:
		return &limiter{}
	case max == 0:
		// Enough to keep every core busy while one request waits on its
		// origin, without letting an unbounded number pile up.
		max = runtime.NumCPU() * 2
	}

	return &limiter{slots: make(chan struct{}, max), wait: wait}
}

// acquire takes a processing slot, queueing briefly so a short burst is
// smoothed rather than refused. It returns a release function, which is only
// meaningful when the error is nil.
func (l *limiter) acquire() (func(), error) {
	if l.slots == nil {
		return func() {}, nil
	}

	release := func() { <-l.slots }

	select {
	case l.slots <- struct{}{}:
		return release, nil
	default:
	}

	timer := time.NewTimer(l.wait)
	defer timer.Stop()

	select {
	case l.slots <- struct{}{}:
		return release, nil
	case <-timer.C:
		return nil, core.NewStatusError(503, "the service is busy, try again shortly")
	}
}

var defaultLimiter = sync.OnceValue(func() *limiter {
	config := core.ReadConfig()

	return newLimiter(config.MaxConcurrent, time.Duration(config.MaxConcurrentWait)*time.Millisecond)
})
