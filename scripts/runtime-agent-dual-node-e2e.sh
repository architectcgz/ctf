#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/code/backend"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ctf-runtime-agent-e2e.XXXXXX")"

SESSION_ID="ctf-e2e-$RANDOM-$$"
DIND_A_NAME="${SESSION_ID}-dind-a"
DIND_B_NAME="${SESSION_ID}-dind-b"

DIND_IMAGE="${DIND_IMAGE:-docker:27-dind}"
DIND_A_PORT="${DIND_A_PORT:-23751}"
DIND_B_PORT="${DIND_B_PORT:-23752}"
AGENT_A_PORT="${AGENT_A_PORT:-19443}"
AGENT_B_PORT="${AGENT_B_PORT:-19444}"
E2E_IMAGE="${E2E_IMAGE:-nginx:1.27-alpine}"

AGENT_A_SERVER_NAME="runtime-agent-a.internal"
AGENT_B_SERVER_NAME="runtime-agent-b.internal"

AGENT_A_LOG="$WORK_DIR/agent-a.log"
AGENT_B_LOG="$WORK_DIR/agent-b.log"
BUILD_LOG="$WORK_DIR/build.log"

cleanup() {
  local status=$?
  trap - EXIT
  set +e

  if [[ $status -ne 0 ]]; then
    echo "[runtime-agent-e2e] failure detected, dumping logs" >&2
    [[ -f "$AGENT_A_LOG" ]] && { echo "--- agent-a.log ---" >&2; cat "$AGENT_A_LOG" >&2; }
    [[ -f "$AGENT_B_LOG" ]] && { echo "--- agent-b.log ---" >&2; cat "$AGENT_B_LOG" >&2; }
    docker logs "$DIND_A_NAME" >&2 2>/dev/null || true
    docker logs "$DIND_B_NAME" >&2 2>/dev/null || true
  fi

  if [[ -n "${AGENT_A_PID:-}" ]] && kill -0 "$AGENT_A_PID" 2>/dev/null; then
    kill "$AGENT_A_PID" 2>/dev/null || true
    wait "$AGENT_A_PID" 2>/dev/null || true
  fi
  if [[ -n "${AGENT_B_PID:-}" ]] && kill -0 "$AGENT_B_PID" 2>/dev/null; then
    kill "$AGENT_B_PID" 2>/dev/null || true
    wait "$AGENT_B_PID" 2>/dev/null || true
  fi

  docker rm -f "$DIND_A_NAME" "$DIND_B_NAME" >/dev/null 2>&1 || true
  rm -rf "$WORK_DIR"
  exit "$status"
}
trap cleanup EXIT

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

wait_for_docker() {
  local host="$1"
  for _ in $(seq 1 60); do
    if docker -H "$host" version >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "docker daemon at $host did not become ready in time" >&2
  return 1
}

wait_for_agent() {
  local endpoint="$1"
  local server_name="$2"
  for _ in $(seq 1 40); do
    if timeout 3s openssl s_client \
      -connect "$endpoint" \
      -servername "$server_name" \
      -CAfile "$WORK_DIR/certs/ca.pem" \
      -cert "$WORK_DIR/certs/client.pem" \
      -key "$WORK_DIR/certs/client-key.pem" \
      -quiet </dev/null >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "runtime-agent at $endpoint did not become ready in time" >&2
  return 1
}

create_leaf_cert() {
  local name="$1"
  local common_name="$2"
  local usage="$3"
  local san="$4"
  local config_file="$WORK_DIR/certs/${name}.cnf"
  local csr_file="$WORK_DIR/certs/${name}.csr"

  cat >"$config_file" <<EOF
[req]
distinguished_name = req_dn
prompt = no
req_extensions = v3_req

[req_dn]
CN = ${common_name}

[v3_req]
keyUsage = digitalSignature,keyEncipherment
extendedKeyUsage = ${usage}
subjectAltName = ${san}
EOF

  openssl req -new -nodes -newkey rsa:2048 \
    -keyout "$WORK_DIR/certs/${name}-key.pem" \
    -out "$csr_file" \
    -config "$config_file" >/dev/null 2>&1

  openssl x509 -req \
    -in "$csr_file" \
    -CA "$WORK_DIR/certs/ca.pem" \
    -CAkey "$WORK_DIR/certs/ca-key.pem" \
    -CAcreateserial \
    -out "$WORK_DIR/certs/${name}.pem" \
    -days 1 \
    -sha256 \
    -extensions v3_req \
    -extfile "$config_file" >/dev/null 2>&1
}

start_runtime_agent() {
  local docker_host="$1"
  local port="$2"
  local server_cert="$3"
  local server_key="$4"
  local log_file="$5"

  (
    cd "$BACKEND_DIR"
    env \
      APP_ENV=dev \
      DOCKER_HOST="$docker_host" \
      CTF_RUNTIME_AGENT_SERVER_ENABLED=true \
      CTF_RUNTIME_AGENT_SERVER_HOST=127.0.0.1 \
      CTF_RUNTIME_AGENT_SERVER_PORT="$port" \
      CTF_RUNTIME_AGENT_SERVER_CERT_FILE="$server_cert" \
      CTF_RUNTIME_AGENT_SERVER_KEY_FILE="$server_key" \
      CTF_RUNTIME_AGENT_SERVER_CLIENT_CA_FILE="$WORK_DIR/certs/ca.pem" \
      CTF_RUNTIME_AGENT_SERVER_SHUTDOWN_TIMEOUT=5s \
      "$WORK_DIR/runtime-agent"
  ) >"$log_file" 2>&1 &
  echo $!
}

require_cmd docker
require_cmd go
require_cmd openssl
require_cmd timeout

mkdir -p "$WORK_DIR/certs"

echo "[runtime-agent-e2e] generating temporary mTLS certificates"
openssl req -x509 -nodes -newkey rsa:2048 \
  -keyout "$WORK_DIR/certs/ca-key.pem" \
  -out "$WORK_DIR/certs/ca.pem" \
  -days 1 \
  -subj "/CN=ctf-runtime-agent-e2e-ca" >/dev/null 2>&1

create_leaf_cert "server-a" "$AGENT_A_SERVER_NAME" "serverAuth" "DNS:${AGENT_A_SERVER_NAME}"
create_leaf_cert "server-b" "$AGENT_B_SERVER_NAME" "serverAuth" "DNS:${AGENT_B_SERVER_NAME}"
create_leaf_cert "client" "runtime-agent-client" "clientAuth" "DNS:runtime-agent-client"

echo "[runtime-agent-e2e] building runtime-agent binary"
(
  cd "$BACKEND_DIR"
  go build -o "$WORK_DIR/runtime-agent" ./cmd/runtime-agent
) >"$BUILD_LOG" 2>&1

echo "[runtime-agent-e2e] starting isolated docker daemons"
docker run -d --privileged \
  --name "$DIND_A_NAME" \
  -p "127.0.0.1:${DIND_A_PORT}:2375" \
  -e DOCKER_TLS_CERTDIR= \
  "$DIND_IMAGE" \
  --host=tcp://0.0.0.0:2375 \
  --host=unix:///var/run/docker.sock \
  --tls=false \
  --storage-driver=vfs >/dev/null

docker run -d --privileged \
  --name "$DIND_B_NAME" \
  -p "127.0.0.1:${DIND_B_PORT}:2375" \
  -e DOCKER_TLS_CERTDIR= \
  "$DIND_IMAGE" \
  --host=tcp://0.0.0.0:2375 \
  --host=unix:///var/run/docker.sock \
  --tls=false \
  --storage-driver=vfs >/dev/null

wait_for_docker "tcp://127.0.0.1:${DIND_A_PORT}"
wait_for_docker "tcp://127.0.0.1:${DIND_B_PORT}"

echo "[runtime-agent-e2e] starting runtime agents"
AGENT_A_PID="$(start_runtime_agent "tcp://127.0.0.1:${DIND_A_PORT}" "$AGENT_A_PORT" "$WORK_DIR/certs/server-a.pem" "$WORK_DIR/certs/server-a-key.pem" "$AGENT_A_LOG")"
AGENT_B_PID="$(start_runtime_agent "tcp://127.0.0.1:${DIND_B_PORT}" "$AGENT_B_PORT" "$WORK_DIR/certs/server-b.pem" "$WORK_DIR/certs/server-b-key.pem" "$AGENT_B_LOG")"

wait_for_agent "127.0.0.1:${AGENT_A_PORT}" "$AGENT_A_SERVER_NAME"
wait_for_agent "127.0.0.1:${AGENT_B_PORT}" "$AGENT_B_SERVER_NAME"

echo "[runtime-agent-e2e] running dual-node cleanup e2e tests"
(
  cd "$BACKEND_DIR"
  env \
    CTF_RUNTIME_ROUTER_E2E_AGENT_A_ENDPOINT="127.0.0.1:${AGENT_A_PORT}" \
    CTF_RUNTIME_ROUTER_E2E_AGENT_A_SERVER_NAME="$AGENT_A_SERVER_NAME" \
    CTF_RUNTIME_ROUTER_E2E_AGENT_B_ENDPOINT="127.0.0.1:${AGENT_B_PORT}" \
    CTF_RUNTIME_ROUTER_E2E_AGENT_B_SERVER_NAME="$AGENT_B_SERVER_NAME" \
    CTF_RUNTIME_ROUTER_E2E_CA_FILE="$WORK_DIR/certs/ca.pem" \
    CTF_RUNTIME_ROUTER_E2E_CLIENT_CERT_FILE="$WORK_DIR/certs/client.pem" \
    CTF_RUNTIME_ROUTER_E2E_CLIENT_KEY_FILE="$WORK_DIR/certs/client-key.pem" \
    CTF_RUNTIME_ROUTER_E2E_DOCKER_HOST_A="tcp://127.0.0.1:${DIND_A_PORT}" \
    CTF_RUNTIME_ROUTER_E2E_DOCKER_HOST_B="tcp://127.0.0.1:${DIND_B_PORT}" \
    CTF_RUNTIME_ROUTER_E2E_IMAGE="$E2E_IMAGE" \
    timeout 900s go test ./internal/app/composition \
      -run 'TestRuntimeNodeExecutionRouterE2ECleanupUsesRuntimeDetailsContainerNode|TestRuntimeNodeExecutionRouterE2ECleanupUsesWorkspaceContainerNode' \
      -count=1 -v
)

echo "[runtime-agent-e2e] success"
