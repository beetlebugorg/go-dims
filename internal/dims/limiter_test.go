package dims

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func TestLimiterSizing(t *testing.T) {
	require.Nil(t, newLimiter(-1, time.Second).slots, "a negative value removes the limit")
	require.Equal(t, 3, cap(newLimiter(3, time.Second).slots))
	require.Equal(t, runtime.NumCPU()*2, cap(newLimiter(0, time.Second).slots),
		"zero derives from the CPU count")
}

func TestLimiterWithoutALimitNeverBlocks(t *testing.T) {
	unlimited := newLimiter(-1, time.Nanosecond)

	for i := 0; i < 100; i++ {
		release, err := unlimited.acquire()
		require.NoError(t, err)
		release()
	}
}

// A request beyond the limit waits, and is refused with 503 if nothing frees up.
func TestLimiterRefusesWhenFull(t *testing.T) {
	full := newLimiter(1, 20*time.Millisecond)

	release, err := full.acquire()
	require.NoError(t, err)

	_, err = full.acquire()
	require.Error(t, err)

	var statusError *core.StatusError
	require.True(t, errors.As(err, &statusError))
	require.Equal(t, 503, statusError.StatusCode)

	// Once the slot frees, the next caller proceeds.
	release()

	second, err := full.acquire()
	require.NoError(t, err)
	second()
}

// A short burst should queue rather than be refused.
func TestLimiterQueuesRatherThanRefusing(t *testing.T) {
	queued := newLimiter(1, time.Second)

	release, err := queued.acquire()
	require.NoError(t, err)

	go func() {
		time.Sleep(30 * time.Millisecond)
		release()
	}()

	start := time.Now()
	second, err := queued.acquire()
	require.NoError(t, err, "a caller must wait for a slot rather than be refused straight away")
	require.GreaterOrEqual(t, time.Since(start), 25*time.Millisecond)
	second()
}

// The cap must actually hold under concurrent use.
func TestLimiterHoldsTheCap(t *testing.T) {
	const limit = 4

	held := newLimiter(limit, 2*time.Second)

	var inFlight, peak int64
	var wg sync.WaitGroup

	for i := 0; i < 40; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			release, err := held.acquire()
			if err != nil {
				return
			}
			defer release()

			current := atomic.AddInt64(&inFlight, 1)
			for {
				highest := atomic.LoadInt64(&peak)
				if current <= highest || atomic.CompareAndSwapInt64(&peak, highest, current) {
					break
				}
			}

			time.Sleep(time.Millisecond)
			atomic.AddInt64(&inFlight, -1)
		}()
	}

	wg.Wait()

	require.LessOrEqual(t, peak, int64(limit), "never more than the limit at once")
	require.Greater(t, peak, int64(1), "the test must have exercised concurrency")
}
