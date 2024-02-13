FROM ghcr.io/beetlebugorg/go-dims:build-latest as build-go-dims

COPY . /build/go-dims
WORKDIR /build/go-dims
RUN go build -o /usr/local/imagemagick/bin/dims cmd/dims/main.go

FROM debian:bookworm-slim

ENV DIMS_DOWNLOAD_TIMEOUT=60000
ENV DIMS_IMAGEMAGICK_TIMEOUT=20000
ENV DIMS_PLACEHOLDER_IMAGE_URL="http://placehold.it/350x150"
ENV DIMS_CACHE_CONTROL_MAX_AGE=604800
ENV DIMS_EDGE_CONTROL_DOWNSTREAM_TTL=604800
ENV DIMS_TRUST_SOURCE=true
ENV DIMS_MIN_SRC_CACHE_CONTROL=604800
ENV DIMS_MAX_SRC_CACHE_CONTROL=604800
ENV DIMS_SECRET_KEY=""
ENV DIMS_CACHE_EXPIRE=604800
ENV DIMS_NO_IMAGE_CACHE_EXPIRE=60
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