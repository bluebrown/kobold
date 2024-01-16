#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

gitea admin user create --username dev \
  --password dev123 --email dev@gitea.local

curl -fsSL "http://localhost:3000/api/v1/user/repos" \
  -u "dev:dev123" -H "Content-Type: application/json" \
  -d '{"name": "test", "default_branch": "main"}'

d=$(mktemp -d)
trap 'rm -rf "$d"' EXIT
tar -xzf /tmp/repo.tar.gz -C "$d"
cd "$d"

git config --global init.defaultBranch main

git init
git config user.email "dev@local"
git config user.name "dev"

git add .
git commit -m "chore: first commit"

git remote add origin http://dev:dev123@localhost:3000/dev/test.git
git push --all origin
