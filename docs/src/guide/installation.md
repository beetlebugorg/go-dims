# Getting Started

## Running with Docker

For most use cases it's probably best to stick with the containers to avoid managing Imagemagick.
Many Linux distributions only distribute Imagemagick 6, and go-dims requires Imagemagick 7.

Execute the following to run in development mode on port 8080:

```shell
$ docker run -e DIMS_SIGNING_KEY=devmode -p 8080:8080 ghcr.io/beetlebugorg/go-dims serve --dev
```

Development mode will disable signature verification so the signing key doesn't really matter, but
it's still required.

You should have dims running now:

```shell
❯ curl http://127.0.0.1:8080/dims-status
ALIVE
```

If everything is working you should see a thumbnail below (after refreshing). If you see alt text,
make sure dims is started.

<img src="http://127.0.0.1:8080/dims4/default/1/1/resize/100x100/?url=https://images.pexels.com/photos/1539116/pexels-photo-1539116.jpeg" alt="[If you have dims running you should see a 100x100 thumbnail.]"/>

## Compiling from source

If you tried to compile mod-dims, go-dims' predecessor, circa 2009 or so it was painful. Things got
better over time with containers and autoconf improvements but it was never _easy_. The process got
a lot easier with go-dims.

You still need to install Imagemagick and its dependencies to compile go-dims but these days you can
get everything you need to compile it on macOS using brew:

```shell
$ brew install imagemagick
```

On Linux you'll need to find an Imagemagick 7 package or compile it from source. It's not hard to
compile, it just takes a while. 

Then you can compile and run go-dims:

```shell
$ git clone https://github.com/beetlebugorg/go-dims.git
$ cd go-dims
$ go build cmd/dims/main.go
$ DIMS_SIGNING_KEY=devmode ./build/bin/dims serve --dev
```