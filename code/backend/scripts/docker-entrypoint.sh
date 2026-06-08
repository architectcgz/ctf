#!/bin/sh
set -eu

log() {
  printf '%s\n' "$*"
}

auto_migrate_enabled() {
  case "${CTF_AUTO_MIGRATE:-false}" in
    1|true|TRUE|yes|YES|on|ON)
      return 0
      ;;
    0|false|FALSE|no|NO|off|OFF)
      return 1
      ;;
    *)
      log "invalid CTF_AUTO_MIGRATE value: ${CTF_AUTO_MIGRATE}"
      exit 1
      ;;
  esac
}

build_migrate_database_url() {
  if [ -n "${MIGRATE_DATABASE_URL:-}" ]; then
    printf '%s' "${MIGRATE_DATABASE_URL}"
    return
  fi

  printf 'postgres://%s:%s@%s:%s/%s?sslmode=%s' \
    "${CTF_POSTGRES_USERNAME:-postgres}" \
    "${CTF_POSTGRES_PASSWORD:-}" \
    "${CTF_POSTGRES_HOST:-127.0.0.1}" \
    "${CTF_POSTGRES_PORT:-5432}" \
    "${CTF_POSTGRES_DATABASE:-ctf}" \
    "${CTF_POSTGRES_SSL_MODE:-disable}"
}

run_migrations() {
  migrate_database_url="$(build_migrate_database_url)"
  log "ctf-api entrypoint: running database migrations"
  /app/migrate \
    -path "${CTF_MIGRATIONS_PATH:-/app/migrations}" \
    -database "${migrate_database_url}" \
    up
  log "ctf-api entrypoint: database migrations finished"
}

if [ "$#" -eq 0 ]; then
  set -- /app/ctf-api
elif [ "${1#-}" != "$1" ]; then
  set -- /app/ctf-api "$@"
fi

if [ "$1" = "/app/ctf-api" ] && auto_migrate_enabled; then
  run_migrations
fi

exec "$@"
