package geometry

import (
	"fmt"
	"testing"
)

func TestGeometry(t *testing.T) {
	tests := []struct {
		geometry string
		expected Geometry
		success  bool
	}{
		{"100", Geometry{Width: 100}, true},
		{"100%", Geometry{Width: 100, Flags: Flags{WidthPercent: true}}, true},
		{"100x", Geometry{Width: 100}, true},
		{"x100", Geometry{Height: 100}, true},
		{"100%x100", Geometry{Width: 100, Height: 100, Flags: Flags{WidthPercent: true, HeightPercent: false}}, true},
		{"abc123", Geometry{}, false},
		{"+100%", Geometry{}, false},
		{"100x100%+50+50", Geometry{Width: 100, Height: 100, X: 50, Y: 50, Flags: Flags{HeightPercent: true}}, true},
		{"12380x7200%+100+100%", Geometry{Width: 12380, Height: 7200, X: 100, Y: 100, Flags: Flags{HeightPercent: true, OffsetYPercent: true}}, true},
		{"100x100!", Geometry{Width: 100, Height: 100, Flags: Flags{Force: true}}, true},
		{"010192309120391092301923x10293012390123-13", Geometry{}, false},
		{"lakjdflajsdf", Geometry{}, false},
		{"---12-31-31-31", Geometry{}, false},
		{"xxxxxxxxxxxxxxxxx100000", Geometry{}, false},
		{"50x50+10a10", Geometry{}, false},
		{"50%x50%+10+10:", Geometry{}, false},
		{"100x200d+0+0,", Geometry{}, false},
		{"75%x100%-10-20", Geometry{}, false},
		{"300x400!<>", Geometry{}, false},
		{"100%x100%!<>", Geometry{}, false},
		{"-50x50+10+10", Geometry{}, false},
		{"50x-50+10+10", Geometry{}, false},
		{"50x50+10-", Geometry{}, false},
		{"50%x50%!<>", Geometry{}, false},
		{"010192309120391092301923x10293012390123-13", Geometry{}, false},
		{"100x200x300", Geometry{}, false},
	}

	for _, test := range tests {
		t.Run(test.geometry, func(t *testing.T) {
			value := parseGeometry(test.geometry)

			if value != test.expected && test.success {
				t.Errorf("expected %v, got %v => '%s'", test.expected, value, test.geometry)
			}
		})
	}
}

func FuzzParseGeometry(f *testing.F) {
	// Seed the fuzzer with initial test cases
	geometries := []string{
		"100", "100%", "100x", "x100", "100%x100", "abc123", "+100%", "100x100%+50+50",
		"12380x7200%+100+100%", "100x100!", "010192309120391092301923x10293012390123-13",
		"lakjdflajsdf", "---12-31-31-31", "xxxxxxxxxxxxxxxxx100000", "50x50+10a10",
		"50%x50%+10+10:", "100x200d+0+0,", "75%x100%-10-20", "300x400!<>", "100%x100%!<>",
		"-50x50+10+10", "50x-50+10+10", "50x50+10-", "50%x50%!<>", "010192309120391092301923x10293012390123-13",
		"100x200x300",
	}

	for _, geometry := range geometries {
		f.Add(geometry)
	}

	f.Fuzz(func(t *testing.T, geometry string) {
		fmt.Println(geometry)
		// Call the function with the fuzzed input
		_ = parseGeometry(geometry)
	})
}

func BenchmarkParseGeometry(b *testing.B) {
	geometries := []string{
		"100", "100%", "100x", "x100", "100%x100", "abc123", "+100%", "100x100%+50+50",
		"12380x7200%+100+100%", "100x100!", "010192309120391092301923x10293012390123-13",
		"lakjdflajsdf", "---12-31-31-31", "xxxxxxxxxxxxxxxxx100000", "50x50+10a10",
		"50%x50%+10+10:", "100x200d+0+0,", "75%x100%-10-20", "300x400!<>", "100%x100%!<>",
		"-50x50+10+10", "50x-50+10+10", "50x50+10-", "50%x50%!<>", "010192309120391092301923x10293012390123-13",
		"100x200x300",
	}

	for _, geometry := range geometries {
		b.Run(geometry, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseGeometry(geometry)
			}
		})
	}
}
