#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

out="${1:-.artifacts}"
out="${out%/}"
tag="${RELEASE_TAG:-dev}"

script_dir="$(dirname "$(readlink -f "$0")")"
cd "$script_dir/.."

mkdir -p "$out" && rm -rf "${out:?}"/*

td="build/.tmp"
trap 'rm -rf "$td"' EXIT
mkdir -p "$td" && rm -rf "${td:?}"/*

cd "$td"
kustomize create
kustomize edit add resource "../kube"
kustomize edit set image "docker.io/bluebrown/kobold:$tag"
kustomize edit set label app.kubernetes.io/name:kobold
kustomize edit set label "app.kubernetes.io/version:$tag"
cd -
kustomize build "$td" >"$out/manifests.yaml"

docker build -f build/Dockerfile -t "docker.io/bluebrown/kobold:$tag" .
docker save "docker.io/bluebrown/kobold:$tag" >"$out/oci.tar"

GOBIN="$script_dir/../$out" go install ./cmd/...

for f in $out/*; do
  if [[ -x "$f" ]]; then
    tar -C "$out" -czf "$out/kobold-${f##*/}.$(go env GOOS)-$(go env GOARCH).tgz" "${f##*/}"
    rm "$f"
  else
    mv "$f" "$out/kobold-${f##*/}"
  fi
done
