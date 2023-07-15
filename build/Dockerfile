 # base image to build from
FROM golang as src
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .


# build with gitexec
FROM src as gitexec-builder
ARG version="dev"
RUN CGO_ENABLED=0 go install --tags gitexec -ldflags "-w -s -X main.version=$version" ./cmd/server

FROM docker.io/alpine/git:v2.36.3 as gitexec
COPY --from=gitexec-builder /go/bin/server /opt/kobold/
ENTRYPOINT ["/opt/kobold/server"]


# build with gitgo
FROM src as gitgo-builder
ARG version="dev"
RUN go install --tags gitgo -ldflags "-w -s -X main.version=$version" ./cmd/server

FROM docker.io/library/debian:bookworm-slim as gitgo
COPY --from=gitgo-builder /go/bin/server /opt/kobold/
ENTRYPOINT ["/opt/kobold/server"]
