#!/usr/bin/env bash

ARGOCD_VERSION=v2.1.2
ARGOCD_CLI_PATH=${ARGOCD_CLI_PATH:-"/tmp/.reviewapp-operator/argocd"}

mkdir -p $(dirname ${ARGOCD_CLI_PATH})

case "$(uname -m)" in
  "x86_64" ) ARGOCD_ARCH=amd64 ;;
  * ) echo "unknown arch"; exit 1 ;;
esac

[ -f ${ARGOCD_CLI_PATH} ] || curl -sSL -o ${ARGOCD_CLI_PATH} https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_VERSION}/argocd-$(uname -s)-${ARGOCD_ARCH}
chmod +x ${ARGOCD_CLI_PATH}
