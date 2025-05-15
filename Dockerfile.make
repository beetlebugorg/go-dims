# syntax=docker/dockerfile:1.7-labs

FROM ghcr.io/beetlebugorg/go-dims:builder AS go-dims
ARG TARGETARCH
RUN apk add zip
COPY --exclude=build . /build/go-dims
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    cd go-dims && \
    go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go mod download && \
    make lambda && \
    mv build/lambda.zip build/lambda-${TARGETARCH}.zip

FROM scratch AS export
ARG TARGETARCH
COPY --from=go-dims /build/go-dims/build/lambda-${TARGETARCH}.zip /lambda-${TARGETARCH}.zip