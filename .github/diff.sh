#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

go mod tidy
go generate ./...
go fmt ./...
git diff --exit-code
