ARG BASE_IMAGE=docker.io/library/debian:bookworm-slim

# builder
FROM golang:1.20

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

COPY . .

ARG CGO_ENABLED=1 BUILD_TAG=gitgo VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install -v -tags "$BUILD_TAG"  -ldflags "-w -s -X main.version=$VERSION" ./cmd/server


# runner
FROM $BASE_IMAGE

RUN echo 'nonroot:x:65532:65532:nonroot:/nonexistent:/sbin/nologin' >>/etc/passwd

COPY --from=0 /go/bin/server /opt/kobold/

COPY build/config.yaml /opt/kobold/config.yaml

ENTRYPOINT ["/opt/kobold/server"]

USER 65532:65532

