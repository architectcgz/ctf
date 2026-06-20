#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${1:-${OUT_DIR:-${BACKEND_DIR}/bin}}"
GO_BIN="${GO:-go}"
TARGET_OS="${GOOS:-linux}"
TARGET_ARCH="${GOARCH:-amd64}"
TARGET_CGO="${CGO_ENABLED:-0}"
TARGET_LDFLAGS="${LDFLAGS:--s -w}"

mkdir -p "${OUT_DIR}"

echo "building ctf-runtime-agent -> ${OUT_DIR}/ctf-runtime-agent (${TARGET_OS}/${TARGET_ARCH})"
(
  cd "${BACKEND_DIR}"
  CGO_ENABLED="${TARGET_CGO}" GOOS="${TARGET_OS}" GOARCH="${TARGET_ARCH}" \
    "${GO_BIN}" build -trimpath -ldflags="${TARGET_LDFLAGS}" -o "${OUT_DIR}/ctf-runtime-agent" ./cmd/runtime-agent
)
