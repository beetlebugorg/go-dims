ARG ALPINE_VERSION=3.21

# -- Alpine Base
FROM alpine:${ALPINE_VERSION} AS alpine-base

COPY scripts/generate-compiled-sbom.sh /usr/local/bin/generate-compiled-sbom.sh

RUN apk add --no-cache alpine-sdk xz zlib-dev zlib-static

# -- Build libpng
# http://www.libpng.org/pub/png/libpng.html
FROM alpine-base AS libpng

ARG PREFIX=/usr/local/dims/libpng
ARG NAME=libing
ARG PNG_LICENSE="Zlib"
ARG PNG_VERSION=1.6.48
ARG PNG_WEBSITE="http://www.libpng.org/pub/png/libpng.html"
ARG PNG_DOWNLOAD="https://download.sourceforge.net/libpng/libpng16/${PNG_VERSION}/libpng-${PNG_VERSION}.tar.xz"
ARG PNG_CHECKSUM="sha256:46fd06ff37db1db64c0dc288d78a3f5efd23ad9ac41561193f983e20937ece03"

ENV PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
ENV LD_LIBRARY_PATH=${PREFIX}/lib

WORKDIR /build

RUN apk add --no-cache alpine-sdk xz zlib-dev zlib-static

ADD --checksum="${PNG_CHECKSUM}" ${PNG_DOWNLOAD} libpng-${PNG_VERSION}.tar.xz

# Build
RUN tar xvf "libpng-${PNG_VERSION}.tar.xz" && \
    cd "libpng-${PNG_VERSION}" && \
    ./configure --prefix="${PREFIX}" --enable-static && \
    make -j"$(nproc)" && \
    make install

# SBOM
RUN cp "libpng-${PNG_VERSION}/LICENSE" "${PREFIX}" && \
    /usr/local/bin/generate-compiled-sbom.sh \
        --name libpng --version ${PNG_VERSION} --license ${PNG_LICENSE} \
        --download ${PNG_DOWNLOAD} --checksum ${PNG_CHECKSUM} --license_file "LICENSE" \
        --website ${PNG_WEBSITE} > ${PREFIX}/sbom.cdx.json

# -- Build libwebp
# https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html
FROM alpine-base AS libwebp

ARG PREFIX=/usr/local/dims/libwebp
ARG NAME=libwebp
ARG WEBP_LICENSE="BSD"
ARG WEBP_VERSION=1.5.0
ARG WEBP_WEBSITE="https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html"
ARG WEBP_DOWNLOAD="https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-${WEBP_VERSION}.tar.gz"
ARG WEBP_CHECKSUM="sha256:7d6fab70cf844bf6769077bd5d7a74893f8ffd4dfb42861745750c63c2a5c92c"

WORKDIR /build

RUN apk add --no-cache alpine-sdk

ADD --checksum="${WEBP_CHECKSUM}" ${WEBP_DOWNLOAD} libwebp-${WEBP_VERSION}.tar.gz

# Build
RUN tar xzvf libwebp-${WEBP_VERSION}.tar.gz && \
    cd libwebp-${WEBP_VERSION} && \
    ./configure --prefix=${PREFIX} --enable-static && \
    make -j"$(nproc)" && \
    make install

# SBOM
RUN cp "libwebp-${WEBP_VERSION}/COPYING" "${PREFIX}/LICENSE" && \
    /usr/local/bin/generate-compiled-sbom.sh \
        --name ${NAME} --version ${WEBP_VERSION} --license ${WEBP_LICENSE} \
        --download ${WEBP_DOWNLOAD} --checksum ${WEBP_CHECKSUM} --license_file "COPYING" \
        --website ${WEBP_WEBSITE} > ${PREFIX}/sbom.cdx.json

# -- Build libtiff
# https://libtiff.gitlab.io/libtiff/
FROM alpine-base AS libtiff

ARG PREFIX=/usr/local/dims
ARG NAME=libtiff
ARG TIFF_VERSION=4.7.0
ARG TIFF_CHECKSUM="sha256:67160e3457365ab96c5b3286a0903aa6e78bdc44c4bc737d2e486bcecb6ba976"
ARG TIFF_LICENSE="libtiff"
ARG TIFF_WEBSITE="https://libtiff.gitlab.io/libtiff/"
ARG TIFF_DOWNLOAD="https://download.osgeo.org/libtiff/tiff-${TIFF_VERSION}.tar.gz"

WORKDIR /build

RUN apk add --no-cache jpeg-dev libjpeg-turbo-static

COPY --from=libwebp ${PREFIX}/libwebp ${PREFIX}/libwebp

ADD --checksum="${TIFF_CHECKSUM}" ${TIFF_DOWNLOAD} tiff-${TIFF_VERSION}.tar.gz

# Build
RUN tar xzvf tiff-${TIFF_VERSION}.tar.gz && \
    cd tiff-${TIFF_VERSION} && \
    ./configure --prefix=$PREFIX/libtiff --enable-static --disable-cxx \
        --with-webp-include-dir=$PREFIX/libwebp/include \
        --with-webp-lib-dir=$PREFIX/libwebp/lib && \
    make -j"$(nproc)" && \
    make install && \
    rm -rf ${PREFIX}/libtiff/bin ${PREFIX}/libtiff/share

# SBOM
RUN cp "tiff-${TIFF_VERSION}/LICENSE.md" "${PREFIX}/libtiff/LICENSE"
RUN /usr/local/bin/generate-compiled-sbom.sh \
        --name ${NAME} --version ${TIFF_VERSION} --license ${TIFF_LICENSE} \
        --download ${TIFF_DOWNLOAD} --checksum ${TIFF_CHECKSUM} --license_file "LICENSE.md" \
        --website ${TIFF_WEBSITE} > ${PREFIX}/libtiff/sbom.cdx.json

# -- Build glib-2.0
# https://docs.gtk.org/glib/
FROM alpine-base AS glib

ARG PREFIX=/usr/local/dims
ARG NAME=glib-2.0
ARG GLIB_MAJOR_MINOR_VERSION=2.84
ARG GLIB_VERSION=2.84.1
ARG GLIB_CHECKSUM="sha256:2b4bc2ec49611a5fc35f86aca855f2ed0196e69e53092bab6bb73396bf30789a"
ARG GLIB_LICENSE="LGPL-2.1-or-later"
ARG GLIB_WEBSITE="https://docs.gtk.org/glib/"
ARG GLIB_DOWNLOAD="https://download.gnome.org/sources/glib/${GLIB_MAJOR_MINOR_VERSION}/glib-${GLIB_VERSION}.tar.xz"

RUN apk add --no-cache meson py3-pip xz upx

WORKDIR /build

ADD --checksum="${GLIB_CHECKSUM}" ${GLIB_DOWNLOAD} glib-${GLIB_VERSION}.tar.xz

# Build
RUN tar -xvf glib-${GLIB_VERSION}.tar.xz && \
    cd glib-${GLIB_VERSION} && \
    meson setup build --prefix=${PREFIX}/glib-2.0 --default-library static --prefer-static --strip --buildtype release -Dauto_features=disabled && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install && \
    upx --best --lzma ${PREFIX}/glib-2.0/bin/* || true

# SBOM
RUN cp "glib-${GLIB_VERSION}/COPYING" "${PREFIX}/glib-2.0/LICENSE" && \
    /usr/local/bin/generate-compiled-sbom.sh \
        --name ${NAME} --version ${GLIB_VERSION} --license ${GLIB_LICENSE} \
        --download ${GLIB_DOWNLOAD} --checksum ${GLIB_CHECKSUM} --license_file "COPYING" \
        --website ${GLIB_WEBSITE} > ${PREFIX}/glib-2.0/sbom.cdx.json

# -- Build libvips
# https://www.libvips.org/
FROM alpine-base AS libvips

ARG PREFIX=/usr/local/dims
ARG NAME=libvips
ARG VIPS_VERSION=8.16.1
ARG VIPS_CHECKSUM="sha256:d114d7c132ec5b45f116d654e17bb4af84561e3041183cd4bfd79abfb85cf724"
ARG VIPS_LICENSE="LGPL-2.1-or-later"
ARG VIPS_WEBSITE="https://www.libvips.org/"
ARG VIPS_DOWNLOAD="https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.xz"

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

ADD --checksum="${VIPS_CHECKSUM}" ${VIPS_DOWNLOAD} vips-${VIPS_VERSION}.tar.xz

RUN tar -xf vips-${VIPS_VERSION}.tar.xz && \
    cd vips-${VIPS_VERSION} && \
    meson setup build --prefix=${PREFIX}/libvips --default-library static --prefer-static --buildtype release \
        -Dauto_features=disabled -Djpeg=enabled -Dlcms=enabled -Dzlib=enabled \
        -Dpng=enabled -Dtiff=enabled -Dwebp=enabled -Ddeprecated=false && \
    cd build && \
    meson compile -j"$(nproc)" && \
    meson install && \
    rm -rf ${PREFIX}/libvips/bin

# SBOM
RUN cp "vips-${VIPS_VERSION}/LICENSE" "${PREFIX}/libvips/LICENSE" && \
    /usr/local/bin/generate-compiled-sbom.sh \
        --name ${NAME} --version ${VIPS_VERSION} --license ${VIPS_LICENSE} \
        --download ${VIPS_DOWNLOAD} --checksum ${VIPS_CHECKSUM} --license_file "COPYING" \
        --website ${VIPS_WEBSITE} > ${PREFIX}/libvips/sbom.cdx.json

# -- Build base
FROM golang:1.24.2-alpine AS builder

WORKDIR /build

ARG PREFIX=/usr/local/dims

COPY --from=libpng      ${PREFIX}/libpng      ${PREFIX}/libpng
COPY --from=libwebp     ${PREFIX}/libwebp     ${PREFIX}/libwebp
COPY --from=libtiff     ${PREFIX}/libtiff     ${PREFIX}/libtiff
COPY --from=libvips     ${PREFIX}/libvips     ${PREFIX}/libvips
COPY --from=glib        ${PREFIX}/glib-2.0    ${PREFIX}/glib-2.0
COPY scripts/install-air.sh .
COPY scripts/install-syft.sh .
COPY scripts/generate-apk-sbom.sh .

ENV PKG_CONFIG_PATH=${PREFIX}/libwebp/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libpng/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libtiff/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/libvips/lib/pkgconfig
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${PREFIX}/glib-2.0/lib/pkgconfig

ENV LD_LIBRARY_PATH=${PREFIX}/libwebp/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libpng/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libtiff/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/libvips/lib
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PREFIX}/glib-2.0/lib

# Build dependencies
RUN apk add --no-cache \
        jpeg-dev libjpeg-turbo-static \
        lcms2-dev lcms2-static \
        giflib-dev giflib-static \
        bzip2-static \
        libsharpyuv \
        expat-dev expat-static \
        zlib-dev zlib-static \
        make alpine-sdk upx openjdk21-jre-headless \
        ca-certificates tzdata gcompat && \
        update-ca-certificates wget vim && \
        cat install-air.sh | sh -s -- -b $(go env GOPATH)/bin && \
        wget https://www.antlr.org/download/antlr-4.13.2-complete.jar && \
        echo 'java -jar /build/antlr-4.13.2-complete.jar $@' > /usr/local/bin/antlr && \
        chmod +x /usr/local/bin/antlr

# Generate sbom for Alpine apk packages
RUN sh /build/generate-apk-sbom.sh zlib-static expat-static \
        jpeg-dev libjpeg-turbo-static \
        lcms2-dev lcms2-static \
        giflib-dev giflib-static \
        bzip2-static \
        libsharpyuv \
        expat-dev expat-static \
        zlib-dev zlib-static musl > apk.sbom.cdx.json && \
    ls -lh .