#!/usr/bin/env bash
set -euo pipefail

THRESHOLD=${1:-95}
PROFILE=${COVERAGE_FILE:-coverage.out}
TMP_PROFILE=$(mktemp)
trap 'rm -f "$TMP_PROFILE"' EXIT

mapfile -t PACKAGES < <(go list ./...)

declare -a COVER_PKGS=()
declare -a OTHER_PKGS=()
for pkg in "${PACKAGES[@]}"; do
  case "$pkg" in
    */cmd/*)
      OTHER_PKGS+=("$pkg")
      ;;
    *)
      COVER_PKGS+=("$pkg")
      ;;
  esac
done

if [ ${#COVER_PKGS[@]} -eq 0 ]; then
  echo "no packages found for coverage" >&2
  exit 1
fi

: > "$PROFILE"
echo "mode: atomic" > "$PROFILE"

IFS=' ' read -r -a EXTRA_FLAGS <<< "${GO_TEST_FLAGS:-}"

for pkg in "${COVER_PKGS[@]}"; do
  go test "${EXTRA_FLAGS[@]}" -covermode=atomic -coverprofile="$TMP_PROFILE" "$pkg"
  tail -n +2 "$TMP_PROFILE" >> "$PROFILE"
  : > "$TMP_PROFILE"
done

if [ ${#OTHER_PKGS[@]} -gt 0 ]; then
  go test "${EXTRA_FLAGS[@]}" "${OTHER_PKGS[@]}"
fi

total=$(go tool cover -func="$PROFILE" | awk 'END {print $3}')
total=${total%\%}

if ! awk -v cov="$total" -v thr="$THRESHOLD" 'BEGIN {exit (cov+0 >= thr+0 ? 0 : 1)}'; then
  echo "Coverage ${total}% is below threshold ${THRESHOLD}%" >&2
  exit 1
fi

echo "Coverage ${total}% meets threshold ${THRESHOLD}%"
