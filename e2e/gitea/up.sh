#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

name=gitea.local
host=localhost:3000
user=kobold
pass=kobold
repo=kobold-test

docker run -d --rm --name "$name" \
  -p "127.0.0.1:$(echo "$host" | cut -d':' -f2):3000" \
  -e GITEA__security__INSTALL_LOCK=true \
  gitea/gitea:1.19-rootless

sleep 10
docker exec "$name" gitea admin user create \
  --admin --username "$user" --password "$pass" \
  --email "$user@gitea.local"

sleep 10
curl "http://$host/api/v1/user/repos" -u "$user:$pass" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'"$repo"'",
    "description": "",
    "private": false,
    "issue_labels": "",
    "auto_init": false,
    "template": false,
    "gitignores": "",
    "license": "",
    "readme": "",
    "default_branch": "main",
    "trust_model": "default"
  }
'

workdir="$(mktemp -d)"

# trap 'rm -rf "$workdir" && docker stop gitea' EXIT

cp -r "$script_dir/repo" "$workdir/repo"
cd "$workdir/repo"

git init
git add .
git commit -sm "chore: first commit"
git remote add origin "http://$host/$user/$repo.git"
git push "http://$user:$pass@$host/$user/$repo.git"
