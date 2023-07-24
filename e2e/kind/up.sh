#!/usr/bin/env bash
set -o nounset -o errexit -o errtrace -o pipefail

script_abs="$(readlink -f "$0")"
script_dir="$(dirname "$script_abs")"

kind create cluster --name kobold-e2e --config "$script_dir/kind.yaml"
kubectl apply -k "$script_dir/"
docker run --rm -d -p 127.0.0.1:5000:5000 --name registry.local registry:2
docker network connect kind registry.local
sleep 30
kubectl wait pod --selector app.kubernetes.io/component=controller --for condition=Ready --timeout 5m --namespace ingress-nginx
