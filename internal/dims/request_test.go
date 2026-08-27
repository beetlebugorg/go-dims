package dims

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A ratio just above 2 used to select a shrink of 4, which decoded below the
// requested size and left the resize to scale back up.
func TestShrinkFactor(t *testing.T) {
	cases := map[int]int{
		0: 1, 1: 1,
		2: 2, 3: 2,
		4: 4, 5: 4, 7: 4,
		8: 8, 9: 8, 64: 8,
	}

	for ratio, want := range cases {
		require.Equal(t, want, shrinkFactor(ratio), "ratio %d", ratio)
	}
}
