VERSION ?= $(shell git rev-parse --short=8 HEAD)
BUILD_DIR := build
BOOTSTRAP := $(BUILD_DIR)/bootstrap

BINARY := $(BUILD_DIR)/dims
LAMBDA_BINARY := $(BUILD_DIR)/bootstrap

LD_FLAGS = "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}'"
STATIC_LDFLAGS = "-X 'github.com/beetlebugorg/go-dims/internal/dims/core.Version=${VERSION}' -linkmode 'external' -extldflags '-fno-PIC -static -Wl,-z,stack-size=8388608 -lpng -lz -ltiff -lwebp -lwebpmux -lwebpdemux -ljpeg -lbz2 -lexpat -llcms2 -lgomp -lsharpyuv'"

all:
	go generate ./...
	go build -o $(BINARY) -ldflags $(LD_FLAGS) ./cmd/dims

static:
	go generate ./...
	go build -o $(BINARY) -ldflags $(STATIC_LDFLAGS) ./cmd/dims

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
