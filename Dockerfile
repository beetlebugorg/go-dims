FROM golang:1.22.0-bookworm as imagemagick-builder

ARG TIFF_VERSION=4.3.0
ARG WEBP_VERSION=1.2.3
ARG IMAGEMAGICK_VERSION=7.1.1-28
ARG PREFIX=/usr/local/go-dims

RUN apt-get -y update && \
    apt-get install -y --no-install-recommends \
      build-essential ca-certificates wget \
      libxml2-dev libgif-dev libjpeg62-turbo-dev libpng-dev liblcms2-dev libfreetype6-dev \
      libtiff5-dev libwebp7 libwebp-dev ca-certificates

RUN wget https://imagemagick.org/archive/ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz && \
    tar -xf ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz

WORKDIR ImageMagick-${IMAGEMAGICK_VERSION}

ENV PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
ENV LD_LIBRARY_PATH=${PREFIX}/lib
RUN ./configure --disable-static --enable-opencl=no --disable-openmp \
    --disable-dpc --without-threads --disable-installed \
    --with-magick-plus-plus=no --with-modules=no --enable-hdri=no \
    --without-utilities --disable-docs --without-openexr --without-lqr --without-x \
    --without-jbig \
    --with-png=yes --with-jpeg=yes --with-xml=yes --with-webp=yes --with-tiff=yes \
    --prefix=${PREFIX}

RUN make -j"$(nproc)" && make install

#FROM golang:1.22.0-alpine3.19
FROM golang:1.22.0-bookworm as build-go-dims

RUN apt-get -y update && \
    apt-get install -y --no-install-recommends \
    libjpeg62-turbo liblcms2-2 libfreetype6 libxml2 libgif7 libpng16-16 libtiff6 libwebp7 \
    libwebpdemux2 libwebpmux3

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

ENV PKG_CONFIG_PATH=/usr/local/go-dims/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/go-dims/lib

COPY --from=imagemagick-builder /usr/local/go-dims /usr/local/go-dims
COPY . /build/go-dims
WORKDIR /build/go-dims
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache make && cp build/dims /usr/local/go-dims/bin/dims

FROM debian:bookworm-slim

RUN apt-get -y update && \
    apt-get install -y --no-install-recommends \
    libjpeg62-turbo liblcms2-2 libfreetype6 libxml2 libgif7 libpng16-16 libtiff6 libwebp7 \
    libwebpdemux2 libwebpmux3 ca-certificates libssl3 && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENV PKG_CONFIG_PATH=/usr/local/go-dims/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/go-dims/lib

COPY --from=build-go-dims /usr/local/go-dims /usr/local/go-dims

ENTRYPOINT ["/usr/local/go-dims/bin/dims"]
CMD ["serve", "--bind", ":8080"]
EXPOSE 8080

USER 33