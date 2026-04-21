#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SAMPLE_DIR="$ROOT_DIR/docs/contracts/examples/challenge-pack-v1/awd-bank-portal-01/docker"
IMAGE_REF="registry.example.edu/ctf/awd-bank-portal:v1"

echo "building AWD sample image: $IMAGE_REF"
docker build -t "$IMAGE_REF" "$SAMPLE_DIR"
echo "built: $IMAGE_REF"
