#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

mkdir -p bin

if ! test -x bin/mdbook; then
  curl -fsSL https://github.com/rust-lang/mdBook/releases/download/v0.4.14/mdbook-v0.4.14-x86_64-unknown-linux-gnu.tar.gz |
    tar -C bin -xzf -
fi

git worktree add /tmp/gh-pages

git -C /tmp/gh-pages/ update-ref -d refs/heads/gh-pages

mv /tmp/gh-pages/.git /tmp/mygit

bin/mdbook build --dest-dir /tmp/gh-pages/ docs/

mv /tmp/mygit /tmp/gh-pages/.git

git -C /tmp/gh-pages/ add .

git -C /tmp/gh-pages/ commit -m "deploy $(git rev-parse main) to gh-pages"

git -C /tmp/gh-pages/ push --force --set-upstream origin gh-pages

git worktree remove /tmp/gh-pages --force
