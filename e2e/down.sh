#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

kind delete cluster --name kobold-e2e
docker stop registry.local gitea.local
