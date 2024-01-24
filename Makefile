.SILENT:
.PHONY: install check lint test vet generate testinfra dev

prefix := $(CURDIR)/.local

export GOBIN := $(prefix)/bin
export PATH := $(GOBIN):$(PATH)

test:
	go test -cover ./...

lint: $(prefix)/bin/golangci-lint
	golangci-lint run --timeout 5m

generate: $(prefix)/bin/sqlc $(prefix)/bin/swag
	go generate ./...

artifacts:
	bash build/artifacts.sh

diff: generate
	git diff --exit-code

check: diff
	$(MAKE) lint test

testinfra: $(prefix)/bin/skaffold
	bash -x e2e/kind/up.sh
	skaffold run -f e2e/skaffold.yaml -p testinfra

dev: $(prefix)/bin/skaffold
	skaffold run -f e2e/skaffold.yaml -p kobold

tools: $(prefix)/bin/golangci-lint $(prefix)/bin/sqlc $(prefix)/bin/skaffold $(prefix)/bin/swag $(prefix)/bin/sqlite3

$(prefix)/bin/skaffold:
	mkdir -p $(prefix)/bin
	curl -fsSL https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 \
	| install /dev/stdin $(prefix)/bin/skaffold

$(prefix)/bin/golangci-lint:
	mkdir -p $(prefix)/bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
	| sh -s -- -b $(prefix)/bin v1.55.2

$(prefix)/bin/sqlc:
	mkdir -p $(prefix)/bin
	curl -fsSL https://downloads.sqlc.dev/sqlc_1.25.0_linux_amd64.tar.gz \
	| tar -C $(prefix)/bin -xzf - sqlc

$(prefix)/bin/swag:
	mkdir -p $(prefix)/bin
	curl -fsSL https://github.com/swaggo/swag/releases/download/v1.16.2/swag_1.16.2_Linux_x86_64.tar.gz \
	| tar -C $(prefix)/bin -xzf - swag

$(prefix)/bin/sqlite3:
	bash build/sqlite3.sh $(prefix)
