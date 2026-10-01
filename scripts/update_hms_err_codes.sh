#!/usr/bin/env sh

set -e

SCRIPT_DIR="$(dirname "$(realpath "${0}")")"
ROOT_DIR="${SCRIPT_DIR}/.."
cd "${ROOT_DIR}"

go run "${SCRIPT_DIR}/hms.go"
echo "Done!"
