#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

bash "$script_dir/kind/up.sh"
bash "$script_dir/gitea/up.sh"
docker network connect kind gitea.local
bash "$script_dir/kobold/up.sh"
