all:
	go build -o ./build/dims go-dims.go

docs:
	mdbook build docs

docs-serve:
	mdbook serve docs
