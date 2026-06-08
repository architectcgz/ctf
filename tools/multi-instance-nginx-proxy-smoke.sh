#!/usr/bin/env bash
set -euo pipefail

if [ "${BASH_SOURCE[0]}" != "$0" ]; then
  echo "multi-instance nginx proxy smoke must be executed, not sourced" >&2
  return 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/code/backend"

POSTGRES_CONTAINER="${CTF_SMOKE_POSTGRES_CONTAINER:-ctf-postgres}"
POSTGRES_DB="${CTF_SMOKE_POSTGRES_DB:-ctf}"
POSTGRES_USER="${CTF_SMOKE_POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${CTF_SMOKE_POSTGRES_PASSWORD:-postgres123456}"

REDIS_CONTAINER="${CTF_SMOKE_REDIS_CONTAINER:-ctf-redis}"
REDIS_PASSWORD="${CTF_SMOKE_REDIS_PASSWORD:-redis123456}"
REDIS_DB="${CTF_SMOKE_REDIS_DB:-15}"

DOCKER_NETWORK="${CTF_SMOKE_DOCKER_NETWORK:-ctf-network}"
NGINX_IMAGE="${CTF_SMOKE_NGINX_IMAGE:-nginx:alpine}"
CURL_IMAGE="${CTF_SMOKE_CURL_IMAGE:-curlimages/curl:latest}"
API_IMAGE="${CTF_SMOKE_API_IMAGE:-ctf-api-nginx-proxy-smoke:$(git -C "$ROOT_DIR" rev-parse --short HEAD)}"
API_COUNT="${CTF_SMOKE_API_COUNT:-3}"
BUILD_TIMEOUT="${CTF_SMOKE_DOCKER_BUILD_TIMEOUT:-300s}"
STARTUP_TIMEOUT_SECONDS="${CTF_SMOKE_STARTUP_TIMEOUT_SECONDS:-45}"
RAW_HTTP_TIMEOUT_SECONDS="${CTF_SMOKE_RAW_HTTP_TIMEOUT_SECONDS:-20}"
KEEP_IMAGE="${CTF_SMOKE_KEEP_IMAGE:-0}"

if [ "$API_COUNT" -lt 2 ]; then
  echo "CTF_SMOKE_API_COUNT must be at least 2" >&2
  exit 2
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "required command not found: $1" >&2
    exit 2
  fi
}

require_cmd docker
require_cmd git
require_cmd python3

TMP_DIR="$(mktemp -d /tmp/ctf-nginx-proxy-smoke.XXXXXX)"
NGINX_CONF="$TMP_DIR/nginx.conf"
TARGET_CONF="$TMP_DIR/target-nginx.conf"
NGINX_LOG_DIR="$TMP_DIR/nginx-logs"
mkdir -p "$NGINX_LOG_DIR"

INSTANCE_ID=$((900000000000 + RANDOM * 1000 + RANDOM))
USER_ID=$((INSTANCE_ID + 1))
CHALLENGE_ID=$((INSTANCE_ID + 2))
TICKET="smoke-ticket-${INSTANCE_ID}"
TARGET_CONTAINER="ctf-proxy-target-smoke-${INSTANCE_ID}"
NGINX_CONTAINER="ctf-nginx-proxy-smoke-${INSTANCE_ID}"
API_CONTAINERS=()

psql_exec() {
  docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" "$POSTGRES_CONTAINER" \
    psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -q -c "$1"
}

redis_cli() {
  docker exec "$REDIS_CONTAINER" redis-cli -a "$REDIS_PASSWORD" -n "$REDIS_DB" "$@"
}

cleanup() {
  status=$?

  docker rm -f "$NGINX_CONTAINER" "$TARGET_CONTAINER" >/dev/null 2>&1 || true
  for container in "${API_CONTAINERS[@]}"; do
    docker rm -f "$container" >/dev/null 2>&1 || true
  done

  redis_cli DEL "ctf:instance:proxy:ticket:${TICKET}" >/dev/null 2>&1 || true
  psql_exec "DELETE FROM public.instances WHERE id = ${INSTANCE_ID}; DELETE FROM public.challenges WHERE id = ${CHALLENGE_ID}; DELETE FROM public.users WHERE id = ${USER_ID};" >/dev/null 2>&1 || true

  if [ "$KEEP_IMAGE" != "1" ]; then
    docker image rm "$API_IMAGE" >/dev/null 2>&1 || true
  fi

  if [ "$status" -eq 0 ]; then
    rm -rf "$TMP_DIR"
  else
    echo "smoke_failed_artifacts=$TMP_DIR" >&2
  fi
}
trap cleanup EXIT INT TERM

assert_docker_prerequisites() {
  docker network inspect "$DOCKER_NETWORK" >/dev/null
  docker inspect "$POSTGRES_CONTAINER" >/dev/null
  docker inspect "$REDIS_CONTAINER" >/dev/null
  if ! docker image inspect "$NGINX_IMAGE" >/dev/null 2>&1; then
    docker pull "$NGINX_IMAGE" >/dev/null
  fi
  if ! docker image inspect "$CURL_IMAGE" >/dev/null 2>&1; then
    docker pull "$CURL_IMAGE" >/dev/null
  fi
}

wait_container_http_ok() {
  local container="$1"
  local url="$2"
  local deadline=$((SECONDS + STARTUP_TIMEOUT_SECONDS))
  while [ "$SECONDS" -lt "$deadline" ]; do
    if docker exec "$container" wget -q -T 2 -O - "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done
  echo "timed out waiting for $container $url" >&2
  docker logs "$container" --tail 120 >&2 || true
  return 1
}

assert_container_ready_status() {
  local container="$1"
  local body
  body="$(docker exec "$container" wget -q -T 2 -O - http://127.0.0.1:8080/ready)"
  python3 - "$body" <<'PY'
import json
import sys

body = sys.argv[1]
payload = json.loads(body)
status = payload.get("data", {}).get("status")
if status != "ready":
    raise SystemExit(f"expected data.status='ready', got {status!r}; body={body}")
PY
}

nginx_raw_get() {
  local path="$1"
  local cookie="${2:-}"
  if [ -n "$cookie" ]; then
    docker run --rm --network "$DOCKER_NETWORK" "$CURL_IMAGE" \
      -sS --http1.1 --max-time "$RAW_HTTP_TIMEOUT_SECONDS" -i \
      -H "Cookie: $cookie" \
      "http://${NGINX_CONTAINER}:8080${path}"
    return
  fi

  docker run --rm --network "$DOCKER_NETWORK" "$CURL_IMAGE" \
    -sS --http1.1 --max-time "$RAW_HTTP_TIMEOUT_SECONDS" -i \
    "http://${NGINX_CONTAINER}:8080${path}"
}

http_status_from_raw_response() {
  python3 -c 'import re, sys; raw = sys.stdin.buffer.read().decode("latin1", errors="replace"); match = re.match(r"HTTP/\S+\s+(\d+)", raw); print(match.group(1) if match else "")'
}

wait_nginx_live() {
  local deadline=$((SECONDS + STARTUP_TIMEOUT_SECONDS))
  local response
  while [ "$SECONDS" -lt "$deadline" ]; do
    response="$(nginx_raw_get "/live" || true)"
    if [ "$(printf '%s' "$response" | http_status_from_raw_response)" = "200" ]; then
      return 0
    fi
    sleep 0.5
  done
  echo "timed out waiting for nginx /live" >&2
  docker logs "$NGINX_CONTAINER" --tail 120 >&2 || true
  [ -f "$NGINX_LOG_DIR/error.log" ] && tail -n 120 "$NGINX_LOG_DIR/error.log" >&2
  return 1
}

write_nginx_conf() {
  {
    cat <<'EOF'
events {}
http {
  log_format upstream_smoke '$request_method $uri $status upstream=$upstream_addr';
  access_log /tmp/ctf-nginx/access.log upstream_smoke;
  error_log /tmp/ctf-nginx/error.log notice;

  upstream ctf_api_smoke {
EOF
    for container in "${API_CONTAINERS[@]}"; do
      printf '    server %s:8080;\n' "$container"
    done
    cat <<'EOF'
  }

  server {
    listen 0.0.0.0:8080;

    location / {
      proxy_pass http://ctf_api_smoke;
      proxy_http_version 1.1;
      proxy_set_header Connection close;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
}
EOF
  } >"$NGINX_CONF"
}

start_target_container() {
  cat >"$TARGET_CONF" <<'EOF'
events {}
http {
  server {
    listen 0.0.0.0:8080;
    location / {
      add_header Content-Type text/plain;
      return 200 'target-ok';
    }
  }
}
EOF
  docker run -d --name "$TARGET_CONTAINER" --network "$DOCKER_NETWORK" \
    -v "$TARGET_CONF:/etc/nginx/nginx.conf:ro" \
    "$NGINX_IMAGE" >/dev/null
  wait_container_http_ok "$TARGET_CONTAINER" "http://127.0.0.1:8080/"
}

build_api_image() {
  timeout "$BUILD_TIMEOUT" docker build -q -f "$BACKEND_DIR/Dockerfile" -t "$API_IMAGE" "$BACKEND_DIR" >/dev/null
}

start_api_containers() {
  local idx
  for idx in $(seq 1 "$API_COUNT"); do
    local name="ctf-api-nginx-smoke-${INSTANCE_ID}-${idx}"
    API_CONTAINERS+=("$name")
    docker run -d --name "$name" --network "$DOCKER_NETWORK" \
      -e APP_ENV=dev \
      -e CTF_APP_ENV=dev \
      -e CTF_HTTP_HOST=0.0.0.0 \
      -e CTF_HTTP_PORT=8080 \
      -e CTF_POSTGRES_HOST=ctf-postgres \
      -e CTF_POSTGRES_PORT=5432 \
      -e CTF_POSTGRES_DATABASE="$POSTGRES_DB" \
      -e CTF_POSTGRES_USERNAME="$POSTGRES_USER" \
      -e CTF_POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
      -e CTF_POSTGRES_SSL_MODE=disable \
      -e CTF_REDIS_ADDR=ctf-redis:6379 \
      -e CTF_REDIS_PASSWORD="$REDIS_PASSWORD" \
      -e CTF_REDIS_DB="$REDIS_DB" \
      -e CTF_CONTAINER_FLAG_GLOBAL_SECRET=0123456789abcdef0123456789abcdef \
      -e CTF_CONTAINER_SCHEDULER_ENABLED=false \
      -e CTF_CHALLENGE_PUBLISH_CHECK_ENABLED=false \
      "$API_IMAGE" >/dev/null
  done

  for container in "${API_CONTAINERS[@]}"; do
    wait_container_http_ok "$container" "http://127.0.0.1:8080/live"
    assert_container_ready_status "$container"
  done
}

seed_proxy_state() {
  local now
  now="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

  psql_exec "
INSERT INTO public.users (id, username, password_hash, role, status, created_at, updated_at)
VALUES (${USER_ID}, 'nginx_proxy_smoke_${INSTANCE_ID}', 'smoke', 'student', 'active', now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.challenges (id, title, description, category, difficulty, points, image_id, status, created_at, updated_at, flag_type, instance_sharing, target_protocol, target_port)
VALUES (${CHALLENGE_ID}, 'nginx proxy smoke', '', 'smoke', 'easy', 1, 0, 'published', now(), now(), 'static', 'shared', 'http', 80)
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.instances (id, user_id, challenge_id, container_id, status, access_url, expires_at, extend_count, max_extends, created_at, updated_at, runtime_details, share_scope)
VALUES (${INSTANCE_ID}, ${USER_ID}, ${CHALLENGE_ID}, '${TARGET_CONTAINER}', 'running', 'http://${TARGET_CONTAINER}:8080', now() + interval '15 minutes', 0, 2, now(), now(), '', 'shared')
ON CONFLICT (id) DO UPDATE SET access_url = EXCLUDED.access_url, status = 'running', expires_at = EXCLUDED.expires_at, updated_at = now();
"

  local claims
  claims="$(python3 - "$USER_ID" "$INSTANCE_ID" "$now" <<'PY'
import json
import sys

user_id = int(sys.argv[1])
instance_id = int(sys.argv[2])
issued_at = sys.argv[3]
print(json.dumps({
    "user_id": user_id,
    "username": f"nginx_proxy_smoke_{instance_id}",
    "role": "student",
    "instance_id": instance_id,
    "share_scope": "shared",
    "purpose": "instance_access",
    "issued_at": issued_at,
}, separators=(",", ":")))
PY
)"
  redis_cli SET "ctf:instance:proxy:ticket:${TICKET}" "$claims" EX 900 >/dev/null
}

start_nginx_container() {
  write_nginx_conf
  docker run -d --name "$NGINX_CONTAINER" --network "$DOCKER_NETWORK" \
    -v "$NGINX_CONF:/etc/nginx/nginx.conf:ro" \
    -v "$NGINX_LOG_DIR:/tmp/ctf-nginx" \
    "$NGINX_IMAGE" >/dev/null
  wait_nginx_live
  docker exec "$NGINX_CONTAINER" sh -c ': > /tmp/ctf-nginx/access.log'
}

assert_proxy_through_nginx() {
  local bootstrap_response="$TMP_DIR/bootstrap.response"
  local proxy_response="$TMP_DIR/proxy.response"

  nginx_raw_get "/api/v1/instances/${INSTANCE_ID}/proxy/check?ticket=${TICKET}&from=bootstrap" >"$bootstrap_response"
  local bootstrap_status
  bootstrap_status="$(http_status_from_raw_response <"$bootstrap_response")"
  if [ "$bootstrap_status" != "302" ]; then
    echo "expected bootstrap proxy request to return 302, got $bootstrap_status" >&2
    cat "$bootstrap_response" >&2 || true
    exit 1
  fi
  if ! tr -d '\r' <"$bootstrap_response" | grep -qi '^Set-Cookie: ctf_instance_proxy_ticket='; then
    echo "expected proxy bootstrap response to set ctf_instance_proxy_ticket" >&2
    cat "$bootstrap_response" >&2 || true
    exit 1
  fi
  if tr -d '\r' <"$bootstrap_response" | grep -qi 'Location: .*ticket='; then
    echo "expected sanitized redirect location without ticket" >&2
    cat "$bootstrap_response" >&2 || true
    exit 1
  fi

  nginx_raw_get "/api/v1/instances/${INSTANCE_ID}/proxy/check?via=cookie" "ctf_instance_proxy_ticket=${TICKET}" >"$proxy_response"
  local proxy_status
  proxy_status="$(http_status_from_raw_response <"$proxy_response")"
  if [ "$proxy_status" != "200" ]; then
    echo "expected cookie proxy request to return 200, got $proxy_status" >&2
    cat "$proxy_response" >&2 || true
    exit 1
  fi
  if ! grep -q 'target-ok' "$proxy_response"; then
    echo "expected target server body through API reverse proxy" >&2
    cat "$proxy_response" >&2 || true
    exit 1
  fi

  mapfile -t PROXY_UPSTREAMS < <(grep "/api/v1/instances/${INSTANCE_ID}/proxy" "$NGINX_LOG_DIR/access.log" | sed -E 's/.*upstream=([^ ]+).*/\1/')
  if [ "${#PROXY_UPSTREAMS[@]}" -lt 2 ]; then
    echo "expected nginx access log to contain both proxy requests" >&2
    cat "$NGINX_LOG_DIR/access.log" >&2 || true
    exit 1
  fi
  if [ "${PROXY_UPSTREAMS[0]}" = "${PROXY_UPSTREAMS[1]}" ]; then
    echo "expected bootstrap and cookie proxy requests to hit different API upstreams, got ${PROXY_UPSTREAMS[0]}" >&2
    cat "$NGINX_LOG_DIR/access.log" >&2 || true
    exit 1
  fi

  printf 'api_containers=%s\n' "${API_CONTAINERS[*]}"
  printf 'nginx_container=%s\n' "$NGINX_CONTAINER"
  printf 'target_container=%s\n' "$TARGET_CONTAINER"
  printf 'instance_id=%s\n' "$INSTANCE_ID"
  printf 'bootstrap_upstream=%s\n' "${PROXY_UPSTREAMS[0]}"
  printf 'cookie_proxy_upstream=%s\n' "${PROXY_UPSTREAMS[1]}"
  printf 'proxy_body=%s\n' "$(grep -o 'target-ok' "$proxy_response" | head -n 1)"
}

assert_docker_prerequisites
build_api_image
start_target_container
start_api_containers
seed_proxy_state
start_nginx_container
assert_proxy_through_nginx
