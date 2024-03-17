# syntax=docker/dockerfile:1.7-labs
ARG ALPINE_VERSION=3.19

# -- Alpine Base
FROM alpine:${ALPINE_VERSION} as alpine-base

RUN apk add --no-cache alpine-sdk xz zlib-dev zlib-static

# -- Build libpng
FROM alpine-base as libpng

ARG PREFIX=/usr/local/dims/libpng
ARG PNG_VERSION=1.6.43
ARG PNG_HASH="sha256:6a5ca0652392a2d7c9db2ae5b40210843c0bbc081cbd410825ab00cc59f14a6c"

ENV PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
ENV LD_LIBRARY_PATH=${PREFIX}/lib

WORKDIR /build

RUN apk add --no-cache alpine-sdk xz zlib-dev zlib-static

ADD --checksum="${PNG_HASH}" \
    https://versaweb.dl.sourceforge.net/project/libpng/libpng16/${PNG_VERSION}/libpng-${PNG_VERSION}.tar.xz \
    libpng-${PNG_VERSION}.tar.xz

RUN tar xvf "libpng-${PNG_VERSION}.tar.xz" && \
    cd "libpng-${PNG_VERSION}" && \
    ./configure --prefix="${PREFIX}" --enable-static && \
    make -j"$(nproc)" && \
    make install

# -- Build libwebp
FROM alpine-base as libwebp

ARG PREFIX=/usr/local/dims/libwebp
ARG WEBP_VERSION=1.2.1
ARG WEBP_HASH="sha256:808b98d2f5b84e9b27fdef6c5372dac769c3bda4502febbfa5031bd3c4d7d018"

WORKDIR /build

RUN apk add --no-cache alpine-sdk

ADD --checksum="${WEBP_HASH}" \
    https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-${WEBP_VERSION}.tar.gz \
    libwebp-${WEBP_VERSION}.tar.gz

RUN tar xzvf libwebp-${WEBP_VERSION}.tar.gz && \
    cd libwebp-${WEBP_VERSION} && \
    ./configure --prefix=${PREFIX} --enable-static && \
    make -j"$(nproc)" && \
    make install

# -- Build libtiff
FROM alpine-base as libtiff

ARG PREFIX=/usr/local/dims
ARG TIFF_VERSION=4.3.0
ARG TIFF_HASH="sha256:0e46e5acb087ce7d1ac53cf4f56a09b221537fc86dfc5daaad1c2e89e1b37ac8"

WORKDIR /build

RUN apk add --no-cache jpeg-dev libjpeg-turbo-static

COPY --from=libwebp ${PREFIX}/libwebp ${PREFIX}/libwebp

ADD --checksum="${TIFF_HASH}" \
    https://download.osgeo.org/libtiff/tiff-${TIFF_VERSION}.tar.gz \
    tiff-${TIFF_VERSION}.tar.gz

RUN tar xzvf tiff-${TIFF_VERSION}.tar.gz && \
    cd tiff-${TIFF_VERSION} && \
    ./configure --prefix=$PREFIX/libtiff --enable-static \
        --with-webp-include-dir=$PREFIX/libwebp/include \
        --with-webp-lib-dir=$PREFIX/libwebp/lib && \
    make -j"$(nproc)" && \
    make install

# -- Build Imagemagick
FROM alpine-base as imagemagick

ARG PREFIX=/usr/local/dims
ARG IMAGEMAGICK_VERSION=7.1.1-29
ARG IMAGEMAGICK_HASH="sha256:f140465fbeb0b4724cba4394bc6f6fb32715731c1c62572d586f4f1c8b9b0685"

WORKDIR /build

RUN apk add --no-cache jpeg-dev libjpeg-turbo-static lcms2-dev lcms2-static bzip2-static

COPY --from=libwebp ${PREFIX}/libwebp ${PREFIX}/libwebp
COPY --from=libtiff ${PREFIX}/libtiff ${PREFIX}/libtiff
COPY --from=libpng  ${PREFIX}/libpng  ${PREFIX}/libpng

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig

ADD --checksum="${IMAGEMAGICK_HASH}" \
    https://imagemagick.org/archive/releases/ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz .

RUN tar xvf ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz && \
    cd ImageMagick-${IMAGEMAGICK_VERSION} && \
    ./configure --enable-opencl --with-openmp --with-magick-plus-plus=no \
    --with-modules=no --enable-hdri=no --without-utilities --disable-dpc \
    --enable-zero-configuration --with-threads --with-quantum-depth=8 \
    --disable-docs --without-openexr --without-lqr --without-x --without-jbig \
    --with-png=yes --with-jpeg=yes --with-xml=yes --with-webp=yes --with-tiff=yes \
    --prefix=${PREFIX}/imagemagick && \
    make -j"$(nproc)" && \
    make install && \
    rm -rf ${PREFIX}/imagemagick/bin && \
    rm -rf ${PREFIX}/imagemagick/etc && \
    rm -rf ${PREFIX}/imagemagick/share

# -- Build glib-2.0
FROM alpine-base as glib

ARG PREFIX=/usr/local/dims
ARG GLIB_MAJOR_MINOR_VERSION=2.80
ARG GLIB_VERSION=2.80.0
ARG GLIB_HASH="sha256:8228a92f92a412160b139ae68b6345bd28f24434a7b5af150ebe21ff587a561d"

RUN apk add --no-cache meson py3-pip xz

WORKDIR /build

ADD --checksum="${GLIB_HASH}" \
    "https://download.gnome.org/sources/glib/2.80/glib-2.80.0.tar.xz" \
    glib-${GLIB_VERSION}.tar.xz

RUN tar -xvf glib-${GLIB_VERSION}.tar.xz && \
    cd glib-${GLIB_VERSION} && \
    meson setup build --prefix=${PREFIX}/glib-2.0 --default-library static --prefer-static \
        -Dauto_features=disabled -Dbuildtype=release && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install

# -- Build libvips
FROM alpine-base as libvips

ARG PREFIX=/usr/local/dims
ARG VIPS_VERSION=8.15.2
ARG VIPS_HASH="sha256:a2ab15946776ca7721d11cae3215f20f1f097b370ff580cd44fc0f19387aee84"

WORKDIR /build

RUN apk add --no-cache \
        jpeg-dev libjpeg-turbo-static \
        lcms2-dev lcms2-static \
        bzip2-static \
        expat-dev expat-static \
        meson py3-pip

COPY --from=libwebp ${PREFIX}/libwebp ${PREFIX}/libwebp
COPY --from=libtiff ${PREFIX}/libtiff ${PREFIX}/libtiff
COPY --from=libpng  ${PREFIX}/libpng  ${PREFIX}/libpng
COPY --from=glib  ${PREFIX}/glib-2.0  ${PREFIX}/glib-2.0

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/glib-2.0/lib/pkgconfig

ADD --checksum="${VIPS_HASH}" \
    https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.xz \
    vips-${VIPS_VERSION}.tar.xz

RUN tar -xf vips-${VIPS_VERSION}.tar.xz && \
    cd vips-${VIPS_VERSION} && \
    meson setup build --prefix=${PREFIX}/libvips --default-library static --prefer-static \
        -Dauto_features=disabled -Djpeg=enabled -Dlcms=enabled -Dzlib=enabled \
        -Dpng=enabled -Dtiff=enabled -Dwebp=enabled -Ddeprecated=false && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install

# -- Build base
FROM golang:1.22.0-alpine as base

WORKDIR /build

ARG PREFIX=/usr/local/dims

COPY --from=libpng      ${PREFIX}/libpng      ${PREFIX}/libpng
COPY --from=libwebp     ${PREFIX}/libwebp     ${PREFIX}/libwebp
COPY --from=libtiff     ${PREFIX}/libtiff     ${PREFIX}/libtiff
COPY --from=imagemagick ${PREFIX}/imagemagick ${PREFIX}/imagemagick
COPY --from=libvips     ${PREFIX}/libvips     ${PREFIX}/libvips
COPY --from=glib        ${PREFIX}/glib-2.0     ${PREFIX}/glib-2.0

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/imagemagick/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libvips/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/glib-2.0/lib/pkgconfig

RUN apk add --no-cache \
        jpeg-dev libjpeg-turbo-static \
        lcms2-dev lcms2-static \
        bzip2-static \
        expat-dev expat-static \
        zlib-dev zlib-static \
        make alpine-sdk upx

# -- Build go-dims

FROM base as go-dims

ENV USER=dims
ENV UID=10001

COPY --exclude="Dockerfile" --exclude="docker/" . /build/go-dims

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apk add --no-cache ca-certificates tzdata gcompat freetype fontconfig && \
    update-ca-certificates

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    cd go-dims && \
    go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go mod download && \
    make static && \
    strip build/dims && \
    upx build/dims

# -- Final
FROM scratch

COPY --from=go-dims /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-dims /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-dims /build/go-dims/build/dims /dims
COPY --from=go-dims /etc/passwd /etc/passwd
COPY --from=go-dims /etc/group /etc/group
COPY --from=go-dims --chown=10001:10001 /tmp /tmp

ENTRYPOINT ["/dims"]
CMD ["serve"]
EXPOSE 8080
USER 10001:10001
