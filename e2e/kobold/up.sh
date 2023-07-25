#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

export CONTAINER_REGISTRY=localhost:5000
export TAG=dev

docker buildx bake -f "$script_dir/../../build/docker-bake.hcl" --push

kubectl apply -k "$script_dir/"
