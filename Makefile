VERSION ?= $(shell git rev-parse --short=8 HEAD)
BUILD_DIR := build
BOOTSTRAP := $(BUILD_DIR)/bootstrap
REGISTRY  := ghcr.io/beetlebugorg/go-dims

BINARY := $(BUILD_DIR)/dims
LAMBDA_BINARY := $(BUILD_DIR)/bootstrap

LD_FLAGS = "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}'"
STATIC_LDFLAGS = "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}' -linkmode 'external' -extldflags '-fno-PIC -static -Wl,-z,stack-size=8388608 -lpng -lz -ltiff -lwebp -lwebpmux -lwebpdemux -ljpeg -lbz2 -lexpat -llcms2 -lgomp -lsharpyuv'"

# -- Build targets

# Builds the shared library version of dims
all:
	go generate ./...
	go build -o $(BINARY) -ldflags $(LD_FLAGS) ./cmd/dims

# Builds a static binary version of dims
static:
	go generate ./...
	go build -o $(BINARY) -ldflags $(STATIC_LDFLAGS) ./cmd/dims

#-- Lambda targets

lambda:
	go generate ./...
	go build -o $(LAMBDA_BINARY) -tags "lambda.norpc lambda" -ldflags $(STATIC_LDFLAGS) ./cmd/dims
	upx $(BUILD_DIR)/bootstrap
	cd $(BUILD_DIR) && zip lambda.zip bootstrap

lambda-amd64:
	docker buildx build \
      --platform linux/amd64 \
      --build-arg TARGETARCH=amd64 \
      -f Dockerfile.make \
      --output type=local,dest=build .

lambda-arm64:
	docker buildx build \
      --platform linux/arm64 \
      --build-arg TARGETARCH=arm64 \
      -f Dockerfile.make \
      --output type=local,dest=build .

deploy-lambda-arm64: lambda-arm64
	cd terraform && terraform apply -auto-approve

#-- Documentation targets

docs:
	mdbook build docs

docs-serve:
	mdbook serve docs

#-- Dockerfile targets

docker:
	docker build -t ${REGISTRY}:local .
	docker images | grep ${REGISTRY}

docker-run: docker
	docker run -it --rm ${REGISTRY}:local /bin/sh

libpng:
	docker buildx build --target libpng --load -t ${REGISTRY}:libpng -f Dockerfile.builder .

libpng-run: libpng
	docker run -it --rm ${REGISTRY}:libpng /bin/sh

libtiff:
	docker buildx build --target libtiff --load -t ${REGISTRY}:libtiff -f Dockerfile.builder .

libtiff-run: libtiff
	docker run --load -t ${REGISTRY}:libtiff /bin/sh

libwebp:
	docker buildx build --target libwebp --load -t ${REGISTRY}:libwebp -f Dockerfile.builder .

libwebp-run: libwebp
	docker run --load -t ${REGISTRY}:libwebp /bin/sh

glib:
	docker buildx build --target glib --load -t ${REGISTRY}:glib -f Dockerfile.builder .

glib-run: glib
	docker run --load -t ${REGISTRY}:glib /bin/sh

libvips:
	docker buildx build --target libvips --load -t ${REGISTRY}:libvips -f Dockerfile.builder .

libvips-run: libvips
	docker run -it --rm ${REGISTRY}:libvips /bin/sh

builder:
	docker buildx build --push --platform linux/amd64,linux/arm64 -t ${REGISTRY}:builder -f Dockerfile.builder .

builder-local:
	docker buildx build --load -t ${REGISTRY}:builder-local -f Dockerfile.builder .

builder-local-run: builder-local
	docker run -it --rm -v .:/build/go-dims ${REGISTRY}:builder-local /bin/sh

.PHONY: docs docs-serve
