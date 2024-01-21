#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

tag="$(sh build/gettag.sh)"

out="${1:-.artifacts}"
out="${out%/}"

echo "Building artifacts in $out" >&2

script_dir="$(dirname "$(readlink -f "$0")")"
cd "$script_dir/.."

mkdir -p "$out" && rm -rf "${out:?}"/*

td="build/.tmp"
trap 'rm -rf "$td"' EXIT
mkdir -p "$td" && rm -rf "${td:?}"/*

echo "Building manifests..." >&2

cd "$td"
kustomize create
kustomize edit add resource "../kube"
kustomize edit set image "docker.io/bluebrown/kobold:$tag"
kustomize edit set label app.kubernetes.io/name:kobold
cat <<EOF >>kustomization.yaml
labels:
- includeSelectors: false
  pairs:
    app.kubernetes.io/version: $tag
EOF
cd -
kustomize build "$td" >"$out/manifests.yaml"

echo "Building OCI image..." >&2

bash build/image.sh

docker save "docker.io/bluebrown/kobold:$tag" >"$out/oci.tar"

echo "Building binaries..." >&2

GOBIN="$script_dir/../$out" go install ./cmd/...

for f in $out/*; do
  if [[ -x "$f" ]]; then
    tar -C "$out" -czf "$out/kobold-${f##*/}.$(go env GOOS)-$(go env GOARCH).tgz" "${f##*/}"
    rm "$f"
  else
    mv "$f" "$out/kobold-${f##*/}"
  fi
done

echo "Done" >&2
