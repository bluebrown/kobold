ARG base=gitgo


# gitexec base
FROM docker.io/alpine/git:v2.36.3 as base-gitexec


# gitgo base
FROM docker.io/library/debian:bookworm-slim as base-gitgo

ARG DEBIAN_FRONTEND=noninteractive

RUN rm -f /etc/apt/apt.conf.d/docker-clean && \
    echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    apt-get -y update && apt-get -y install ca-certificates

RUN echo 'nonroot:x:65532:65532:nonroot:/nonexistent:/sbin/nologin' >>/etc/passwd


# go builder
FROM golang:1.20 as builder

ARG DEBIAN_FRONTEND=noninteractive

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download -x

COPY . .

ARG CGO_ENABLED=1 BUILD_TAG=gitgo VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go install -v -tags "$BUILD_TAG"  -ldflags "-w -s -X main.version=$VERSION" ./cmd/server


# final
FROM "base-$base" as final

COPY --from=builder /go/bin/server /opt/kobold/

COPY build/config.yaml /opt/kobold/config.yaml

ENTRYPOINT ["/opt/kobold/server"]

USER 65532:65532

