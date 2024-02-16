all:
	go build -o ./build/bin/dims cmd/dims/main.go

docs:
	mdbook build docs


docs-serve:
	mdbook serve docs
