#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
GO_BIN=${GO:-go}
MAKE_BIN=${MAKE:-make}
OUTPUT_DIR=${OUTPUT_DIR:-"$ROOT_DIR/bin"}
BINARY_NAME=${BINARY_NAME:-finance-server}
SKIP_GENERATE=${SKIP_GENERATE:-0}
GOCACHE_DIR=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}

GOPATH_BIN=$("$GO_BIN" env GOPATH)/bin
export PATH="$GOPATH_BIN:$PATH"

mkdir -p "$OUTPUT_DIR" "$GOCACHE_DIR"
export GOCACHE="$GOCACHE_DIR"

if [[ "$SKIP_GENERATE" != "1" ]]; then
	echo "==> running code generation"
	(cd "$ROOT_DIR" && "$MAKE_BIN" generate)
fi

if [[ ! -f "$ROOT_DIR/go.sum" ]]; then
	echo "==> generating go.sum"
	(cd "$ROOT_DIR" && "$GO_BIN" mod tidy)
fi

echo "==> building $BINARY_NAME"
(cd "$ROOT_DIR" && "$GO_BIN" build -o "$OUTPUT_DIR/$BINARY_NAME" ./cmd/server)

echo "==> build complete: $OUTPUT_DIR/$BINARY_NAME"
