#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="$SCRIPT_DIR/ledger"

if [[ ! -x "$BIN_PATH" ]]; then
  echo "ledger binary not found or not executable: $BIN_PATH" >&2
  exit 1
fi

exec "$BIN_PATH" "$@"
