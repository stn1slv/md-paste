#!/usr/bin/env bash
# Assemble the md-paste.app menu bar bundle and a distributable zip.
# Usage: scripts/build-app.sh [version]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${1:-${VERSION:-dev}}"
DIST="$ROOT/dist"
APP="$DIST/md-paste.app"
PKG="github.com/stn1slv/md-paste/internal/cli"
COMMIT="$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo none)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-s -w -X ${PKG}.version=${VERSION} -X ${PKG}.commit=${COMMIT} -X ${PKG}.date=${DATE}"

echo "Building universal binary (version ${VERSION})..."
mkdir -p "$DIST"
AMD64_BIN="$DIST/md-paste-amd64"
ARM64_BIN="$DIST/md-paste-arm64"
# Clean up the intermediate per-arch binaries even if a later step fails.
trap 'rm -f "$AMD64_BIN" "$ARM64_BIN"' EXIT
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$AMD64_BIN" "$ROOT/cmd/md-paste"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$ARM64_BIN" "$ROOT/cmd/md-paste"
echo "Assembling $APP..."
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

# The universal binary is the bundle's main executable (a real Mach-O, so the
# app has a proper code identity). Launched with no arguments (double-click),
# it detects the .app context and starts menu bar mode.
lipo -create -output "$APP/Contents/MacOS/md-paste" "$AMD64_BIN" "$ARM64_BIN"
rm -f "$AMD64_BIN" "$ARM64_BIN"
chmod +x "$APP/Contents/MacOS/md-paste"

sed "s/__VERSION__/${VERSION}/g" "$ROOT/build/Info.plist.tmpl" >"$APP/Contents/Info.plist"

if [ ! -f "$ROOT/build/icon.icns" ]; then
	"$ROOT/scripts/make-icns.sh"
fi
cp "$ROOT/build/icon.icns" "$APP/Contents/Resources/icon.icns"

echo "Ad-hoc signing..."
codesign --force --sign - "$APP"

echo "Zipping..."
ZIP="$DIST/md-paste_${VERSION}_macos.zip"
rm -f "$ZIP"
ditto -c -k --keepParent "$APP" "$ZIP"
echo "Created $ZIP"
