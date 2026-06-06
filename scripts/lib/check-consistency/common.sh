red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

check_file() {
  if [[ -f "$1" ]]; then
    echo "  $(green PASS) — $1"
  else
    echo "  $(red FAIL) — missing $1"
    fail=1
  fi
}

check_dir() {
  if [[ -d "$1" ]]; then
    echo "  $(green PASS) — $1"
  else
    echo "  $(red FAIL) — missing $1"
    fail=1
  fi
}

check_contains() {
  local file="$1" pattern="$2" label="$3"
  if [[ ! -f "$file" ]]; then
    echo "  $(red FAIL) — $label: missing $file"
    fail=1
  elif grep -qE "$pattern" "$file"; then
    echo "  $(green PASS) — $label"
  else
    echo "  $(red FAIL) — $label"
    fail=1
  fi
}

check_not_contains() {
  local file="$1" pattern="$2" label="$3"
  if [[ ! -f "$file" ]]; then
    echo "  $(red FAIL) — $label: missing $file"
    fail=1
  elif grep -qE "$pattern" "$file"; then
    echo "  $(red FAIL) — $label"
    fail=1
  else
    echo "  $(green PASS) — $label"
  fi
}
