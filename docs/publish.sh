#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

cd "$script_dir/../"

git worktree add /tmp/gh-pages

git -C /tmp/gh-pages/ update-ref -d refs/heads/gh-pages

mv /tmp/gh-pages/.git /tmp/mygit

mdbook build --dest-dir /tmp/gh-pages/ docs/

mv /tmp/mygit /tmp/gh-pages/.git

git -C /tmp/gh-pages/ add .

git -C /tmp/gh-pages/ commit -m "deploy $(git rev-parse main) to gh-pages"

git -C /tmp/gh-pages/ push --force --set-upstream origin gh-pages

git worktree remove /tmp/gh-pages --force
