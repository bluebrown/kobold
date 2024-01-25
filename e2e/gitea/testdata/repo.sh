#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_dir="$(dirname "$(readlink -f "$0")")"
cd "$script_dir/repo"

tar -czf ../repo.tar.gz .
