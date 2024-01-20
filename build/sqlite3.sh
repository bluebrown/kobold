#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

prefix="${1:-/usr/local}"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

curl -fsSL "https://www.sqlite.org/2023/sqlite-autoconf-3440200.tar.gz" |
  tar --strip-components 1 -xzf -

autoreconf -fi
CFLAGS="-Os -DSQLITE_ENABLE_LOAD_EXTENSION" ./configure --prefix "$prefix"
make -j"$(nproc)"
make install

mkdir -p "$prefix/lib"

curl -fsSL https://www.sqlite.org/src/raw/4011aef176616872b2a0d5bccf0ecfb1f7ce3fe5c3d107f3a8e949d8e1e3f08d?at=sha1.c -o sha1.c
gcc -shared -fPIC -o "$prefix/lib/sha1.so" sha1.c

curl -fsSL https://www.sqlite.org/src/raw/5bb2264c1b64d163efa46509544fd7500cb8769cb7c16dd52052da8d961505cf?at=uuid.c -o uuid.c
gcc -shared -fPIC -o "$prefix/lib/uuid.so" uuid.c

cat <<EOF >"$prefix/bin/sqlite"
#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail
exec sqlite3 -cmd '.load uuid' -cmd '.load sha1' "\$@"
EOF

chmod +x "$prefix/bin/sqlite"
