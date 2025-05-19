FROM ghcr.io/beetlebugorg/go-dims:builder-local AS go-dims

ARG VERSION="v0.0.0"

ENV USER=dims
ENV UID=10001

COPY . /build/go-dims

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Build
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    cd go-dims && \
    go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go mod download && \
    make static VERSION=${VERSION}

# Generate sbom for go-dims binary
RUN sh go-dims/scripts/install-syft.sh && \
    cd go-dims && \
    /build/bin/syft file:build/dims -o spdx-json > go-dims.sbom.spdx.json

RUN cd go-dims && strip build/dims && upx build/dims

# -- Final
FROM scratch

LABEL org.opencontainers.image.description="On-the-fly dynamic image management server."

COPY --from=go-dims /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-dims /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-dims /etc/passwd /etc/passwd
COPY --from=go-dims /etc/group /etc/group
COPY --from=go-dims --chown=10001:10001 /tmp /tmp

COPY --from=go-dims /build/go-dims/build/dims /dims
COPY --from=go-dims /build/go-dims/LICENSES /LICENSES
COPY --from=go-dims /build/go-dims/LICENSE /LICENSE
COPY --from=go-dims /build/go-dims/NOTICE /NOTICE

# SBOMs
COPY --from=go-dims /build/go-dims/go-dims.sbom.spdx.json /sbom.spdx.json
COPY --from=go-dims /build/apk.sbom.cdx.json /sbom/apk.sbom.cdx.json
COPY --from=go-dims /usr/local/dims/libpng/sbom.cdx.json /sbom/libpng.sbom.cdx.json
COPY --from=go-dims /usr/local/dims/libwebp/sbom.cdx.json /sbom/libwebp.sbom.cdx.json
COPY --from=go-dims /usr/local/dims/libtiff/sbom.cdx.json /sbom/libtiff.sbom.cdx.json
COPY --from=go-dims /usr/local/dims/glib-2.0/sbom.cdx.json /sbom/glib-2.0.sbom.cdx.json
COPY --from=go-dims /usr/local/dims/libvips/sbom.cdx.json /sbom/libvips.sbom.cdx.json

ENV DIMS_LOG_FORMAT=json

HEALTHCHECK --interval=5s --timeout=2s --start-period=5s --retries=3 \
    CMD /dims health-check || exit 1

ENTRYPOINT ["/dims"]
CMD ["serve"]
EXPOSE 8080
USER 10001:10001