SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

BIN_KO = bin/ko
BIN_KUBECTL = bin/kubectl
BIN_KUSTOMIZE = bin/kustomize

KO_CONFIG_PATH = ./hack/
KO_DOCKER_REPO = index.docker.io/bluebrown/kobold
VERSION = v0.1.0

export KO_CONFIG_PATH
export KO_DOCKER_REPO
export VERSION


##@ Options

KUBE_VERSION ?= 1.24.6
TARGET ?= dev
GITHUB_SHA ?= $(shell git rev-parse main)


##@ Commands

.PHONY: help
help: ## Display this help text
	@awk -F '(:.*##|?=)' \
		'BEGIN                  { printf "\n\033[1mUsage:\033[0m\n  make \033[36m[ COMMAND ]\033[0m \33[35m[ OPTION=VALUE ]...\33[0m\n" } \
		/^[A-Z_]+\s\?=\s+.+/    { printf "  \033[35m%-17s\033[0m (default:%s)\n", $$1, $$2 } \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 } \
		/^##@/                  { printf "\n\033[1m%s:\033[0m\n", substr($$0, 5) } \
		' $(MAKEFILE_LIST)

.PHONY: validate
validate:
	go vet ./...
	go test ./...

.PHONY: deploy
deploy: bin/kustomize bin/kubectl ## Deploy kobold into kubernetes using the TARGET kustomization
	@touch "manifests/$(TARGET)/etc/.env" "manifests/$(TARGET)/etc/config.yaml"
	@$(BIN_KUSTOMIZE) build "manifests/$(TARGET)" | $(BIN_KUBECTL) apply -f -

.PHONY: undeploy
undeploy: bin/kustomize bin/kubectl ## Undeploy using the TARGET kustomization
	@$(BIN_KUSTOMIZE) build "manifests/$(TARGET)" | $(BIN_KUBECTL) delete -f -

.PHONY: load
load: bin/ko ## Load the image into a local kind cluster
	@$(BIN_KO) build --bare --push=false --tarball=hack/img.tar --tags=dev ./cmd/server/
	@kind load image-archive hack/img.tar

.PHONY: publish
publish: bin/ko ## Build and push the image
	@$(BIN_KO) build --bare --tags="$(VERSION)" --tags=latest ./cmd/server/

.PHONY: dist
dist: ## Create artifacts to attach to github release
	@mkdir -p .dist
	@$(BIN_KUSTOMIZE) build "manifests/dist" -o .dist/kobold-manifests.yaml
	@go build -ldflags='-w -s -extldflags "-static" -X main.version="$(VERSION)"' -o .dist/kobold -v -race ./cmd/server/
	@tar -czf .dist/kobold-amd64.tar.gz  .dist/kobold
	@$(BIN_KO) build --bare --push=false --tarball=.dist/kobold-image.tar --tags="$(VERSION)" ./cmd/server/
	@rm -rf .dist/kobold

.PHONY: docs
docs: bin/mdbook ## Deploy the docs to github pages using the branch gh-pages
	@git worktree add /tmp/gh-pages
	@git -C /tmp/gh-pages/ update-ref -d refs/heads/gh-pages
	@mv /tmp/gh-pages/.git /tmp/mygit
	@mdbook build --dest-dir /tmp/gh-pages/ docs/
	@mv /tmp/mygit /tmp/gh-pages/.git
	@git -C /tmp/gh-pages/ add .
	@git -C /tmp/gh-pages/ commit -m "deploy $(GITHUB_SHA) to gh-pages"
	@git -C /tmp/gh-pages/ push --force --set-upstream origin gh-pages
	@git worktree remove /tmp/gh-pages --force


# Dependencies

bin/ko:
	@mkdir -p bin
	@curl -fsSL https://github.com/ko-build/ko/releases/download/v0.12.0/ko_0.12.0_Linux_x86_64.tar.gz \
		| tar --exclude='LICENSE' --exclude='README.md' -C bin/ -xzf -
	@chmod +x bin/ko

bin/kubectl:
	@mkdir -p bin
	@curl -fsSL "https://dl.k8s.io/release/v$(KUBE_VERSION)/bin/$(shell go env GOOS)/$(shell go env GOARCH)/kubectl" -o bin/kubectl
	@chmod +x bin/kubectl

bin/kustomize:
	@mkdir -p bin
	@cd bin && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash

bin/mdbook:
	@curl -fsSL https://github.com/rust-lang/mdBook/releases/download/v0.4.14/mdbook-v0.4.14-x86_64-unknown-linux-gnu.tar.gz | tar -C bin -xzf -
