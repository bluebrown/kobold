FROM docker.io/library/alpine:3.19 as base

FROM base as compiler
RUN apk add --no-cache --update autoconf automake libtool make gcc g++ musl-dev

FROM base as downloader
WORKDIR /src/sqlite
RUN apk add --no-cache --update curl
RUN curl -fsSL "https://www.sqlite.org/2023/sqlite-autoconf-3440200.tar.gz" \
    | tar --strip-components 1 -xzf -

FROM compiler as sqlite
WORKDIR /src/sqlite
COPY --from=downloader /src/sqlite /src/sqlite
RUN autoreconf -fi
RUN CFLAGS="-Os -DSQLITE_ENABLE_LOAD_EXTENSION" ./configure --prefix /usr/local
RUN make -j"$(nproc)"
RUN make install
ADD https://www.sqlite.org/src/raw/4011aef176616872b2a0d5bccf0ecfb1f7ce3fe5c3d107f3a8e949d8e1e3f08d?at=sha1.c sha1.c
RUN gcc -shared -fPIC -o /usr/local/lib/sha1.so sha1.c
ADD https://www.sqlite.org/src/raw/5bb2264c1b64d163efa46509544fd7500cb8769cb7c16dd52052da8d961505cf?at=uuid.c uuid.c
RUN gcc -shared -fPIC -o /usr/local/lib/uuid.so uuid.c

FROM base as runtime
RUN apk add --no-cache --update ca-certificates git openssh-client
COPY --from=sqlite /usr/local/bin /usr/local/bin
COPY --from=sqlite /usr/local/lib /usr/local/lib
RUN ldconfig /usr/local/lib
COPY build/.sqliterc /root/

FROM runtime as nonroot
ARG NONROOT_HOME=/etc/kobold
WORKDIR "$NONROOT_HOME"
RUN echo "nonroot:x:65532:65532:nonroot:$NONROOT_HOME:/sbin/nologin" >> /etc/passwd
RUN echo "nonroot:x:65532:" > /etc/group
RUN cp /root/.sqliterc "$NONROOT_HOME"
RUN chown -R 65532:65532 "$NONROOT_HOME"
USER 65532:65532

FROM docker.io/golang:1.21.5-alpine3.19 as builder
WORKDIR /go/src
COPY . .
ARG BUILD_TARGET=server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "/go/bin/run" "./cmd/$BUILD_TARGET"

FROM nonroot as release
COPY --from=builder /go/bin/run /usr/local/bin
ENTRYPOINT ["/usr/local/bin/run"]
ENV GIT_AUTHOR_NAME="kobold[bot]" \
    GIT_AUTHOR_EMAIL="kobold@noreply.local"
ENV GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME" \
    GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL"
