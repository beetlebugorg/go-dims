ARG ALPINE_VERSION=3.21

# -- Alpine Base
FROM alpine:${ALPINE_VERSION} AS alpine-base

RUN apk add --no-cache alpine-sdk xz zlib-dev zlib-static

# -- Build libpng
# http://www.libpng.org/pub/png/libpng.html
FROM alpine-base AS libpng

ARG PREFIX=/usr/local/dims/libpng
ARG PNG_VERSION=1.6.48
ARG PNG_HASH="sha256:46fd06ff37db1db64c0dc288d78a3f5efd23ad9ac41561193f983e20937ece03"

ENV PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
ENV LD_LIBRARY_PATH=${PREFIX}/lib

WORKDIR /build

ADD --checksum="${PNG_HASH}" \
    https://versaweb.dl.sourceforge.net/project/libpng/libpng16/${PNG_VERSION}/libpng-${PNG_VERSION}.tar.xz \
    libpng-${PNG_VERSION}.tar.xz

RUN tar xvf "libpng-${PNG_VERSION}.tar.xz" && \
    cd "libpng-${PNG_VERSION}" && \
    ./configure --prefix="${PREFIX}" --enable-static && \
    make -j"$(nproc)" && \
    make install

# -- Build mozjpeg
# https://github.com/mozilla/mozjpeg
FROM alpine-base AS mozjpeg

ARG PREFIX=/usr/local/dims
ARG MOZJPEG_VERSION=4.1.1
ARG MOZJPEG_HASH="sha256:66b1b8d6b55d263f35f27f55acaaa3234df2a401232de99b6d099e2bb0a9d196"

ENV CMAKE_PREFIX_PATH=/usr/local/dims/libpng
ENV CMAKE_INSTALL_PREFIX=/usr/local/dims/mozjpeg

WORKDIR /build

RUN apk add --no-cache cmake nasm

COPY --from=libpng  ${PREFIX}/libpng  ${PREFIX}/libpng

ADD --checksum="${MOZJPEG_HASH}" \
    https://github.com/mozilla/mozjpeg/archive/refs/tags/v4.1.1.tar.gz \
    mozjpeg-v${MOZJPEG_VERSION}.tar.gz

RUN tar xvf "mozjpeg-v${MOZJPEG_VERSION}.tar.gz" && \
    cd "mozjpeg-${MOZJPEG_VERSION}" && \
    mkdir build && cd build && \
    cmake -G"Unix Makefiles" ../ -DCMAKE_INSTALL_LIBDIR=${CMAKE_INSTALL_PREFIX}/lib && \
    make -j"$(nproc)" && \
    make install && echo ""

# -- Build libwebp
# https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html
FROM alpine-base AS libwebp

ARG PREFIX=/usr/local/dims/libwebp
ARG WEBP_VERSION=1.5.0
ARG WEBP_HASH="sha256:7d6fab70cf844bf6769077bd5d7a74893f8ffd4dfb42861745750c63c2a5c92c"

WORKDIR /build

ADD --checksum="${WEBP_HASH}" \
    https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-${WEBP_VERSION}.tar.gz \
    libwebp-${WEBP_VERSION}.tar.gz

RUN tar xzvf libwebp-${WEBP_VERSION}.tar.gz && \
    cd libwebp-${WEBP_VERSION} && \
    ./configure --prefix=${PREFIX} --enable-static && \
    make -j"$(nproc)" && \
    make install

# -- Build libtiff
# https://libtiff.gitlab.io/libtiff/
FROM alpine-base AS libtiff

ARG PREFIX=/usr/local/dims
ARG TIFF_VERSION=4.7.0
ARG TIFF_HASH="sha256:67160e3457365ab96c5b3286a0903aa6e78bdc44c4bc737d2e486bcecb6ba976"

WORKDIR /build

COPY --from=libwebp ${PREFIX}/libwebp ${PREFIX}/libwebp
COPY --from=mozjpeg ${PREFIX}/mozjpeg ${PREFIX}/mozjpeg

ADD --checksum="${TIFF_HASH}" \
    https://download.osgeo.org/libtiff/tiff-${TIFF_VERSION}.tar.gz \
    tiff-${TIFF_VERSION}.tar.gz

RUN tar xzvf tiff-${TIFF_VERSION}.tar.gz && \
    cd tiff-${TIFF_VERSION} && \
    ./configure --prefix=$PREFIX/libtiff --enable-static --disable-cxx \
        --with-webp-include-dir=$PREFIX/libwebp/include \
        --with-webp-lib-dir=$PREFIX/libwebp/lib \
        --with-jpeg-include-dir=$PREFIX/mozjpeg/include \
        --with-jpeg-lib-dir=$PREFIX/mozjpeg/lib && \
    make -j"$(nproc)" && \
    make install && \
    rm -rf $PREFIX/libtiff/bin $PREFIX/libtiff/share

# -- Build glib-2.0
# https://docs.gtk.org/glib/
FROM alpine-base AS glib

ARG PREFIX=/usr/local/dims
ARG GLIB_MAJOR_MINOR_VERSION=2.84
ARG GLIB_VERSION=2.84.1
ARG GLIB_HASH="sha256:2b4bc2ec49611a5fc35f86aca855f2ed0196e69e53092bab6bb73396bf30789a"

RUN apk add --no-cache meson py3-pip upx

WORKDIR /build

ADD --checksum="${GLIB_HASH}" \
    "https://download.gnome.org/sources/glib/${GLIB_MAJOR_MINOR_VERSION}/glib-${GLIB_VERSION}.tar.xz" \
    glib-${GLIB_VERSION}.tar.xz

RUN tar -xvf glib-${GLIB_VERSION}.tar.xz && \
    cd glib-${GLIB_VERSION} && \
    meson setup build --prefix=${PREFIX}/glib-2.0 --default-library static --prefer-static --strip --buildtype release -Dauto_features=disabled && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install && \
    upx --best --lzma ${PREFIX}/glib-2.0/bin/* || true

# -- Build lcms2
# https://github.com/mm2/Little-CMS
FROM alpine-base AS liblcms2

ARG PREFIX=/usr/local/dims
ARG LCMS_VERSION=2.17
ARG LCMS_HASH="sha256:d11af569e42a1baa1650d20ad61d12e41af4fead4aa7964a01f93b08b53ab074"

WORKDIR /build

COPY --from=mozjpeg ${PREFIX}/mozjpeg ${PREFIX}/mozjpeg

ADD --checksum="${LCMS_HASH}" \
    "https://github.com/mm2/Little-CMS/releases/download/lcms2.17/lcms2-2.17.tar.gz" \
    lcms2-${LCMS_VERSION}.tar.gz

RUN tar -xvf lcms2-${LCMS_VERSION}.tar.gz && \
    cd lcms2-${LCMS_VERSION} && \
    ./configure --prefix=${PREFIX}/lcms2 \
        --with-jpeg=$PREFIX/mozjpeg \
        --enable-static && \
    make -j"$(nproc)" && \
    make install

# -- Build libvips
# https://www.libvips.org/
FROM alpine-base AS libvips

ARG PREFIX=/usr/local/dims
ARG VIPS_VERSION=8.16.1
ARG VIPS_HASH="sha256:d114d7c132ec5b45f116d654e17bb4af84561e3041183cd4bfd79abfb85cf724"

WORKDIR /build

RUN apk add --no-cache bzip2-static expat-dev expat-static meson py3-pip

COPY --from=libwebp  ${PREFIX}/libwebp  ${PREFIX}/libwebp
COPY --from=libtiff  ${PREFIX}/libtiff  ${PREFIX}/libtiff
COPY --from=libpng   ${PREFIX}/libpng   ${PREFIX}/libpng
COPY --from=glib     ${PREFIX}/glib-2.0 ${PREFIX}/glib-2.0
COPY --from=mozjpeg  ${PREFIX}/mozjpeg  ${PREFIX}/mozjpeg
COPY --from=liblcms2 ${PREFIX}/lcms2    ${PREFIX}/lcms2

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/glib-2.0/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/mozjpeg/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/lcms2/lib/pkgconfig

ADD --checksum="${VIPS_HASH}" \
    https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.xz \
    vips-${VIPS_VERSION}.tar.xz

RUN tar -xf vips-${VIPS_VERSION}.tar.xz && \
    cd vips-${VIPS_VERSION} && \
    meson setup build --prefix=${PREFIX}/libvips --default-library static --prefer-static --buildtype release \
        -Dauto_features=disabled -Djpeg=enabled -Dlcms=enabled -Dzlib=enabled \
        -Dpng=enabled -Dtiff=enabled -Dwebp=enabled -Ddeprecated=false && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install && \
    rm -rf ${PREFIX}/libvips/bin

# -- Build base
FROM golang:1.24.2-alpine

WORKDIR /build

ARG PREFIX=/usr/local/dims

COPY --from=libpng      ${PREFIX}/libpng      ${PREFIX}/libpng
COPY --from=libwebp     ${PREFIX}/libwebp     ${PREFIX}/libwebp
COPY --from=libtiff     ${PREFIX}/libtiff     ${PREFIX}/libtiff
COPY --from=libvips     ${PREFIX}/libvips     ${PREFIX}/libvips
COPY --from=glib        ${PREFIX}/glib-2.0    ${PREFIX}/glib-2.0
COPY --from=mozjpeg     ${PREFIX}/mozjpeg     ${PREFIX}/mozjpeg
COPY --from=liblcms2    ${PREFIX}/lcms2       ${PREFIX}/lcms2

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libvips/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/glib-2.0/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/mozjpeg/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/lcms2/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libwebp/lib/pkgconfig

ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libpng/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libtiff/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libvips/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/glib-2.0/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/mozjpeg/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/lcms2/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libwebp/lib

RUN apk add --no-cache \
        giflib-dev giflib-static \
        bzip2-static \
        libsharpyuv \
        expat-dev expat-static \
        zlib-dev zlib-static \
        make alpine-sdk upx openjdk21-jre-headless \
        ca-certificates tzdata gcompat && \
        update-ca-certificates wget vim && \
        wget https://www.antlr.org/download/antlr-4.13.2-complete.jar && \
        echo 'java -jar /build/antlr-4.13.2-complete.jar $@' > /usr/local/bin/antlr && \
        chmod +x /usr/local/bin/antlr