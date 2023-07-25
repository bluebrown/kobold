#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

if [ "${PRE_RELEASE:-0}" = 1 ]; then
  pre=true
fi

git tag -a -m "release" "$VERSION"
git push "$(git remote show)" "$VERSION"

rid="$(curl -fsSL \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/bluebrown/vscode-extension-yamlfmt/releases \
  -d '{ "tag_name":"'"$VERSION"'", "generate_release_notes": true, "prerelease": '"${pre:-false}"' }' |
  jq -r '.id')"

find "${DIST_DIR:-$script_dir/../.dist/}" -type f |
  while read -r asset_path; do
    curl -fsSL \
      -X POST \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "X-GitHub-Api-Version: 2022-11-28" \
      -H "Content-Type: application/octet-stream" \
      "https://uploads.github.com/repos/bluebrown/kobold/releases/$rid/assets?name=${asset_path##*/}" \
      --data-binary "@$asset_path" >/dev/null
  done
