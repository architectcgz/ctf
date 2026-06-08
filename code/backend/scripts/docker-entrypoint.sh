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

print_legacy_chain_reset_hint() {
  printf '%s\n' \
    "ctf-api entrypoint: this repository now uses a single baseline migration." \
    "ctf-api entrypoint: databases created from the removed 000002..000012 chain are no longer upgraded in place." \
    "ctf-api entrypoint: reset the local database or PostgreSQL volume, then rerun migrations." >&2
}

run_migrations() {
  migrate_database_url="$(build_migrate_database_url)"
  log "ctf-api entrypoint: running database migrations"
  set +e
  migration_output="$(
    /app/migrate \
      -path "${CTF_MIGRATIONS_PATH:-/app/migrations}" \
      -database "${migrate_database_url}" \
      up 2>&1
  )"
  migration_status=$?
  set -e
  if [ "${migration_status}" -ne 0 ]; then
    printf '%s\n' "${migration_output}" >&2
    case "${migration_output}" in
      *"no migration found for version "*)
        print_legacy_chain_reset_hint
        ;;
    esac
    exit "${migration_status}"
  fi
  if [ -n "${migration_output}" ]; then
    printf '%s\n' "${migration_output}"
  fi
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
