#!/usr/bin/env bash
# Generate build/icon.icns from build/icon-1024.png using macOS tooling.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/build/icon-1024.png"

if [ ! -f "$SRC" ]; then
	echo "error: $SRC not found" >&2
	exit 1
fi

TMPROOT="$(mktemp -d)"
trap 'rm -rf "$TMPROOT"' EXIT
ICONSET="$TMPROOT/icon.iconset"
mkdir -p "$ICONSET"
for size in 16 32 128 256 512; do
	double=$((size * 2))
	sips -z "$size" "$size" "$SRC" --out "$ICONSET/icon_${size}x${size}.png" >/dev/null
	sips -z "$double" "$double" "$SRC" --out "$ICONSET/icon_${size}x${size}@2x.png" >/dev/null
done

iconutil -c icns "$ICONSET" -o "$ROOT/build/icon.icns"
echo "Created $ROOT/build/icon.icns"
