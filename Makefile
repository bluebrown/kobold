.SILENT:
.PHONY: install check lint test vet generate testinfra dev
bin=.local/bin

artifacts:
	bash build/artifacts.sh

testinfra: $(bin)/skaffold
	bash -x e2e/kind/up.sh
	skaffold run -f e2e/skaffold.yaml -p testinfra

dev: $(bin)/skaffold
	$(bin)/skaffold run -f e2e/skaffold.yaml -p kobold

$(bin)/skaffold:
	mkdir -p $(bin)
	curl -fsSL https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 \
	| install /dev/stdin $(bin)/skaffold
