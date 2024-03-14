FROM golang:1.22.0-alpine as build-go-dims

ARG PREFIX=/usr/local/imagemagick
ARG IMAGEMAGICK_VERSION=7.1.1-29
ARG WEBP_VERSION=1.2.1
ARG TIFF_VERSION=4.3.0
ARG PNG_VERSION=1.6.43

ENV PKG_CONFIG_PATH=${PREFIX}/lib/pkgconfig
ENV LD_LIBRARY_PATH=${PREFIX}/lib

ENV USER=dims
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /build

RUN apk add --no-cache \
    libjpeg-turbo-dev libjpeg-turbo-static \
    zstd-static zlib-static bzip2-static \
    fontconfig-static freetype-static libxml2-static \
    brotli-static expat-static \
    lcms2-dev lcms2-static \
    librsvg-dev make alpine-sdk wget vim gdb upx ca-certificates tzdata && \
    update-ca-certificates

# -- Build libpng

RUN wget https://versaweb.dl.sourceforge.net/project/libpng/libpng16/${PNG_VERSION}/libpng-${PNG_VERSION}.tar.xz && \
    tar xvf libpng-${PNG_VERSION}.tar.xz && \
    cd libpng-${PNG_VERSION} && \
    ./configure --prefix=${PREFIX} --enable-static && \
    make -j4 && make install

# -- Build webp

RUN wget https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-${WEBP_VERSION}.tar.gz && \
    tar xzvf libwebp-${WEBP_VERSION}.tar.gz && \
    cd libwebp-${WEBP_VERSION} && \
    ./configure --prefix=${PREFIX} --enable-static && \
    make -j4 && make install

# -- Build tiff

RUN wget https://download.osgeo.org/libtiff/tiff-${TIFF_VERSION}.tar.gz && \
    tar xzvf tiff-${TIFF_VERSION}.tar.gz && \
    cd tiff-${TIFF_VERSION} && \
    ./configure --prefix=$PREFIX --enable-static --with-webp-include-dir=$PREFIX/include --with-webp-lib-dir=$PREFIX/lib && \
    make -j4 && make install

# -- Build Imagemagick

RUN wget https://imagemagick.org/archive/releases/ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz && \
    tar -xf ImageMagick-${IMAGEMAGICK_VERSION}.tar.xz && \
    cd ImageMagick-${IMAGEMAGICK_VERSION} && \
    ./configure --enable-opencl --with-openmp --with-magick-plus-plus=no \
    --with-modules=no --enable-hdri=no --without-utilities --disable-dpc \
    --enable-zero-configuration --with-threads --with-quantum-depth=8 \
    --disable-docs --without-openexr --without-lqr --without-x --without-jbig \
    --with-png=yes --with-jpeg=yes --with-xml=yes --with-webp=yes --with-tiff=yes \
    --prefix=${PREFIX} && \
    make -j4 && \
    make install

COPY . /build/go-dims
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    cd go-dims && \
    go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go mod download && \
    make static && \
    strip build/dims && \
    upx build/dims

FROM scratch

COPY --from=build-go-dims /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-go-dims /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go-dims /build/go-dims/build/dims /dims
COPY --from=build-go-dims /etc/passwd /etc/passwd
COPY --from=build-go-dims /etc/group /etc/group

ENTRYPOINT ["/dims"]
EXPOSE 8080
USER 10001:10001

