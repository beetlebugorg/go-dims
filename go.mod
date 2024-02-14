module github.com/beetlebugorg/go-dims

go 1.22.0

require (
	github.com/alecthomas/kong v0.8.1
	gopkg.in/gographics/imagick.v3 v3.5.1
)

replace gopkg.in/gographics/imagick.v3 => ./override/imagick

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/sagikazarmark/slog-shim v0.1.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
)
