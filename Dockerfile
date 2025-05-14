FROM ghcr.io/beetlebugorg/go-dims:builder AS go-dims

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

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    cd go-dims && \
    go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go mod download && \
    make static VERSION=${VERSION} && \
    strip build/dims && \
    upx build/dims

# -- Final
FROM scratch

LABEL org.opencontainers.image.description "On-the-fly dynamic image resizing server."

COPY --from=go-dims /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-dims /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-dims /build/go-dims/build/dims /dims
COPY --from=go-dims /etc/passwd /etc/passwd
COPY --from=go-dims /etc/group /etc/group
COPY --from=go-dims --chown=10001:10001 /tmp /tmp

HEALTHCHECK --interval=5s --timeout=2s --start-period=5s --retries=3 \
    CMD /dims health-check || exit 1

ENTRYPOINT ["/dims"]
CMD ["serve"]
EXPOSE 8080
USER 10001:10001