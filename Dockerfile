FROM ghcr.io/beetlebugorg/go-dims:build-latest as build-go-dims

COPY . /build/go-dims
WORKDIR /build/go-dims
RUN go build -o /usr/local/imagemagick/bin/dims cmd/dims/main.go

FROM debian:bookworm-slim

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
ENV DIMS_DOWNLOAD_TIMEOUT=60000
ENV DIMS_PLACEHOLDER_BACKGROUND="#5adafd"
ENV DIMS_PLACEHOLDER_IMAGE_EXPIRE=60
ENV DIMS_DEFAULT_EXPIRE=31536000
ENV DIMS_STRIP_METADATA=true
ENV DIMS_INCLUDE_DISPOSITION=false
ENV DIMS_CACHE_CONTROL_MAX_AGE=604800
ENV DIMS_EDGE_CONTROL_DOWNSTREAM_TTL=604800
ENV DIMS_TRUST_SRC=true
ENV DIMS_MIN_SRC_CACHE_CONTROL=604800
ENV DIMS_MAX_SRC_CACHE_CONTROL=604800
ENV DIMS_CACHE_EXPIRE=604800

# DIMS_SIGNING_ALGORITHM can be md5 or hmac-sha256
# For compatibility with the original DIMS, set to md5.
ENV DIMS_SIGNING_ALGORITHM=md5

#ENV DIMS_DEFAULT_IMAGE_PREFIX=""
#ENV DIMS_DEFAULT_OUTPUT_FORMAT=webp
#ENV DIMS_IGNORE_DEFAULT_OUTPUT_FORMATS=jpeg,png

ENV LC_ALL="C"

USER root

COPY --from=build-go-dims /usr/local/imagemagick /usr/local/imagemagick

RUN apt-get update && \
    apt-get -y install \
        libpangocairo-1.0-0 libgif7 libjpeg62-turbo libpng16-16 libgomp1 libjbig0 liblcms2-2 \
        libbz2-1.0 libfftw3-double3 libfontconfig1 libfreetype6 libheif1 \
        liblqr-1-0 libltdl7 liblzma5 libopenjp2-7 libopenexr-3-1-30 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENV LD_LIBRARY_PATH=/usr/local/imagemagick/lib

ENTRYPOINT ["/usr/local/imagemagick/bin/dims"]
CMD ["serve", "--bind", ":8080"]
EXPOSE 8080

USER 33