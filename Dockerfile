FROM alpine:edge as build-go-dims

RUN apk add --no-cache \
    imagemagick-dev imagemagick-webp \
    imagemagick-jpeg imagemagick-tiff imagemagick-svg \
    imagemagick-pango \
    go upx make alpine-sdk

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

COPY . /build/go-dims
WORKDIR /build/go-dims
RUN make

FROM alpine:3.19

#-- Imagemagick Settings
# 
# For a complete list of ImageMagick environment variables, see: https://imagemagick.org/script/resources.php
#

# MAGIC_TIME_LIMIT is the maximum time in seconds that the ImageMagick operations will run
ENV MAGICK_TIME_LIMIT=1

# MAGICK_MEMORY_LIMIT is the maximum amount of memory that the ImageMagick operations will use
# before caching to disk.
ENV MAGICK_MEMORY_LIMIT=128MB

#-- DIMS Settings
#
# For a complete list of DIMS environment variables, see: internal/dims/config.go
#

ENV DIMS_SECRET_KEY=""

ENV DIMS_CACHE_CONTROL_USE_ORIGIN=true
ENV DIMS_CACHE_CONTROL_DEFAULT=31536000
ENV DIMS_CACHE_CONTROL_MIN=0
ENV DIMS_CACHE_CONTROL_MAX=31536000
ENV DIMS_CACHE_CONTROL_ERROR=60
ENV DIMS_EDGE_CONTROL_DOWNSTREAM_TTL=604800

#ENV DIMS_DEFAULT_IMAGE_PREFIX=""
#ENV DIMS_DEFAULT_OUTPUT_FORMAT=webp
#ENV DIMS_IGNORE_DEFAULT_OUTPUT_FORMATS=jpeg,png
#ENV DIMS_DOWNLOAD_TIMEOUT=60000
#ENV DIMS_ERROR_BACKGROUND="#5adafd"
#ENV DIMS_STRIP_METADATA=true
#ENV DIMS_INCLUDE_DISPOSITION=false

ENV LC_ALL="C"

USER root

RUN apk add --no-cache \
    imagemagick-webp \
    imagemagick-jpeg imagemagick-tiff imagemagick-svg \
    imagemagick-pango

COPY --from=build-go-dims /build/go-dims/build/dims /usr/local/bin/dims

ENTRYPOINT ["/usr/local/bin/dims"]
CMD ["serve", "--bind", ":8080"]
EXPOSE 8080

USER 33