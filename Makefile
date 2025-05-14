VERSION ?= $(shell git rev-parse --short=8 HEAD)

all:
	go generate ./...
	go build -o ./build/dims -ldflags "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}'" go-dims.go

static:
	go generate ./...
	go build -o ./build/dims -ldflags "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}' -linkmode 'external' -extldflags '-fno-PIC -static -Wl,-z,stack-size=8388608 -lpng -lz -ltiff -lwebp -lwebpmux -lwebpdemux -ljpeg -lbz2 -lexpat -llcms2 -lgomp -lgif -lsharpyuv'" go-dims.go

docs:
	mdbook build docs

docs-serve:
	mdbook serve docs

docker: Dockerfile
	docker buildx build --load -t ghcr.io/beetlebugorg/go-dims:local .
	docker images | grep ghcr.io/beetlebugorg/go-dims

builder:
	docker buildx build --push --platform linux/amd64,linux/arm64 -t ghcr.io/beetlebugorg/go-dims:builder -f Dockerfile.builder .

devmedia:
	docker run --rm --name go-dims-devmedia --privileged -p 8081:80 -v ./resources:/usr/share/nginx/html:ro nginx:latest

.PHONY: docs docs-serve devmedia
