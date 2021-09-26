#!/usr/bin/env bash

ARGOCD_CLI_PATH=${ARGOCD_CLI_PATH:-"/tmp/argocd"}

case "$(uname -m)" in
  "x86_64" ) ARGOCD_ARCH=amd64 ;;
  * ) echo "unknown arch"; exit 1 ;;
esac

curl -sSL -o ${ARGOCD_CLI_PATH} \
  https://github.com/argoproj/argo-cd/releases/latest/download/argocd-$(uname -s)-${ARGOCD_ARCH}

