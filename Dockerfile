FROM ghcr.io/beetlebugorg/go-dims:builder as go-dims

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

RUN apk add --no-cache \
        jpeg-dev libjpeg-turbo-static \
        lcms2-dev lcms2-static \
        bzip2-static \
        expat-dev expat-static \
        zlib-dev zlib-static \
        make alpine-sdk upx \
        ca-certificates tzdata gcompat freetype fontconfig && \
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
