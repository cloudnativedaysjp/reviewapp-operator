#!/usr/bin/env bash

KUSTOMIZE_VERSION=4.4.0
KUSTOMIZE_PATH=${KUSTOMIZE_PATH:-"/tmp/.reviewapp-operator/kustomize"}

mkdir -p $(dirname ${KUSTOMIZE_PATH})

[ -f ${KUSTOMIZE_PATH} ] || bash <(curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh) ${KUSTOMIZE_VERSION} $(dirname ${KUSTOMIZE_PATH})

