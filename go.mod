module github.com/beetlebugorg/go-dims

go 1.22.0

require github.com/alecthomas/kong v0.8.1

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/lipgloss v0.9.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/sagikazarmark/slog-shim v0.1.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	gopkg.in/gographics/imagick.v3 v3.5.1
)

replace gopkg.in/gographics/imagick.v3 => ./override/imagick
