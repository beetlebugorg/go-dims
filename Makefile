all:
	go build -o ./build/dims go-dims.go

static:
	go build -o ./build/dims -ldflags "-linkmode 'external' -extldflags '-fno-PIC -static -Wl,-z,stack-size=8388608 -lpng -lz -ltiff -lzstd -lwebp -lwebpmux -lwebpdemux -ljpeg -lbz2 -lfontconfig -lfreetype -lexpat -lbrotlidec -lbrotlienc -lbrotlicommon -llcms2'" go-dims.go

docs:
	mdbook build docs

docs-serve:
	mdbook serve docs

docker-build: Dockerfile.build
	docker buildx build --load -t beetlebugorg/go-dims:build . -f Dockerfile.build
	docker images | grep beetlebugorg/go-dims

docker: Dockerfile
	docker buildx build --load -t beetlebugorg/go-dims:local .
	docker images | grep beetlebugorg/go-dims

