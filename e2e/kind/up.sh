#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_dir="$(dirname "$(readlink -f "$0")")"

cd "$script_dir"

if ! kind get clusters | grep -q "kobold"; then
  kind create cluster "$@" --config kind.yaml --wait 5m
else
  echo "kind cluster already exists"
fi
