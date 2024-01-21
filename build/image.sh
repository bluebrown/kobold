#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_dir="$(dirname "$(readlink -f "$0")")"
cd "$script_dir/.."

tag="$(sh build/gettag.sh)"
vcs_ref="$(git rev-parse --short HEAD)"

docker build -t "docker.io/bluebrown/kobold:$tag" \
  --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
  --build-arg VCS_REF="$vcs_ref" \
  -f build/Dockerfile .
