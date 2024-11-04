export CGO_CFLAGS="`pkg-config --cflags MagickWand`"
export CGO_LDFLAGS="`pkg-config --libs MagickWand`"
export CGO_CFLAGS_ALLOW='-Xpreprocessor'
