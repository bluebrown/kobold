.SILENT:
.PHONY: install check lint test vet generate testinfra dev

bin=.local/bin
INSTALL_DIR ?= $(CURDIR)/$(bin)

RELEASE_TAG ?= $(shell git describe --tags --always --dirty)

image:
	docker build -t docker.io/bluebrown/kobold:$(RELEASE_TAG) -f build/Dockerfile .

install:
	GOBIN=$(INSTALL_DIR) go install -ldflags '-w -s' ./cmd/...

check:
	go mod tidy
	$(MAKE) generate
	go fmt ./...
	git diff --exit-code
	go vet ./...
	go test ./...

generate:
	go generate ./...

dbshell:
	docker run --rm -ti -v /home/blue/.config/kobold:/tmp \
		--entrypoint sqlite3 kobold /tmp/kobold.sqlite3

testinfra: $(bin)/skaffold
	bash -x e2e/kind/up.sh
	skaffold run -f e2e/skaffold.yaml -p testinfra

dev: $(bin)/skaffold
	$(bin)/skaffold run -f e2e/skaffold.yaml -p kobold

$(bin)/skaffold:
	mkdir -p $(bin)
	curl -fsSL https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 \
	| install /dev/stdin $(bin)/skaffold

