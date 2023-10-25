SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

PATH := $(CURDIR)/bin:$(PATH)
export PATH

GOBIN = $(CURDIR)/bin/
export GOBIN

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)

##@ Options

BUILD_TAGS ?= gitgo gitexec
VERSION ?= dev
PRE_RELEASE ?= 0
CONTAINER_REGISTRY ?= docker.io
BUILDX_FLAGS ?= --load
DIST_DIR ?= .dist

export VERSION
export PRE_RELEASE
export CONTAINER_REGISTRY

remote = $(shell git remote show)
head = $(shell git remote show $(remote) | sed -n '/HEAD branch/s/.*: //p')
current = $(shell git branch --show-current)


##@ Commands

.PHONY: help
help: bin/makehelp ## Show this help text
	bin/makehelp Makefile


###@ Develop

.PHONY: e2e-up
e2e-up: bin/kind bin/kubectl bin/kustomize ## Spin up the local e2e setup
	bash e2e/up.sh

.PHONY: e2e-down
e2e-down: bin/kind bin/kubectl bin/kustomize ## Tear down the local e2e setup
	bash e2e/down.sh

e2e-reload: ## Reload kobold
	CONTAINER_REGISTRY=localhost:5000 $(MAKE) image-publish
	kubectl rollout restart deploy/kobold -n kobold


###@ Validate

.PHONY: check
check: license-check vet test ## Run all checks

.PHONY: vet
vet: ## Lint the code
	go vet ./...

.PHONY: test
test: ## Run test suite
	go test ./...

.PHONY: license-check
license-check: bin/go-licenses ## Check for dangerous licenses of dependencies
	go-licenses check --include_tests  \
	--allowed_licenses=0BSD,ISC,BSD-2-Clause,BSD-3-Clause,MIT,Apache-2.0,MPL-2.0 \
	./cmd/server/

.PHONY: git-ishead
git-ishead: # Fail, if current branch is not HEAD
	test $(current) = $(head)

.PHONY: git-isclean
git-isclean: # Fail, if worktree is dirty
	git diff-index --quiet HEAD --


###@ Build

.PHONY: build
build: | bin ## Build the binaries with for each tag of BUILD_TAGS
	$(foreach tag,$(BUILD_TAGS),go build -tags $(tag) \
		-ldflags='-w -s -X "main.version=$(VERSION)"' -o bin/kobold$(if $(filter-out $(tag), gitgo),-$(tag)) ./cmd/server/;)

.PHONY: image-build
image-build: ## Build the images with VERSION as tag. Passes BUILDX_FLAGS to buildx
	docker buildx bake --file build/docker-bake.hcl $(BUILDX_FLAGS)

.PHONY: artifacts
artifacts: ## Create all release artifacts and put the in .dist/
	mkdir -p .dist && rm -rf .dist/*
	cp assets/schema.json .dist/kobold.schema.json
	$(MAKE) artifact-binary artifact-image artifact-manifests

artifact-binary: build
	$(foreach binary,$(wildcard bin/kobold*),tar -C bin -czf .dist/$(notdir $(binary)).$(GOOS)-$(GOARCH).tgz $(notdir $(binary));)

artifact-image:
	$(MAKE) image-build BUILDX_FLAGS='--set *.attest=type=sbom \
		--set gitgo.output=type=tar,dest=.dist/kobold.oci.tar \
		--set gitexec.output=type=tar,dest=.dist/kobold-gitexec.oci.tar'

artifact-manifests: bin/kustomize
	bin/kustomize build manifests/base/ \
		| sed -E 's|(index.docker.io/bluebrown/kobold):.+|\1:$(VERSION)|g' \
		| sed -E 's|(app.kubernetes.io/version):.+|\1: $(VERSION)|g' \
		>.dist/kobold.manifests.yaml


###@ Publish

.PHONY: version-next
version-next: # internal command to set VERSION to the next semver and IS_LATEST accordingly
	$(if $(filter $(PRE_RELEASE), 0), $(eval IS_LATEST = 1))
	$(eval VERSION = v$(shell docker run --rm -u "$$(id -u):$$(id -g)" \
		-v $(CURDIR):/tmp -w /tmp convco/convco version --bump \
		$(if $(filter $(PRE_RELEASE), 1),--prerelease rc)))

.PHONY: image-publish
image-publish: ## Build and push the images to CONTAINER_REGISTRY
	IS_LATEST=$(IS_LATEST) $(MAKE) image-build BUILDX_FLAGS='--set *.attest=type=sbom --set=*.output=type=registry'

.PHONY: github-pages
github-pages: bin/mdbook ## Build and publish the docs to github pages
	bash docs/publish.sh

.PHONY: github-release
github-release: git-ishead git-isclean version-next ## Create a new release on GitHub and publish the images. Set PRE_RELEASE=1 for pre releases
	$(MAKE) artifacts
	bash .github/release.sh
	$(MAKE) image-publish


## Dependencies

bin:
	mkdir -p bin

bin/makehelp: | bin
	curl -fsSL https://gist.githubusercontent.com/bluebrown/2ec155902622b5e46e2bfcbaff342eb9/raw/Makehelp.awk | install /dev/stdin bin/makehelp

bin/kubectl: | bin
	curl -fsSL "https://dl.k8s.io/release/$(shell curl -fsSL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" -o bin/kubectl

bin/kustomize: | bin
	curl -fsSL "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s bin

bin/kind: | bin
	curl -fsSL https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64 | install /dev/stdin bin/kind

bin/mdbook: | bin
	curl -fsSL https://github.com/rust-lang/mdBook/releases/download/v0.4.32/mdbook-v0.4.32-x86_64-unknown-linux-gnu.tar.gz | tar -C bin -xzf -

bin/go-licenses: | bin
	go install github.com/google/go-licenses@latest
