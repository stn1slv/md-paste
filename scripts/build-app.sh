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
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST/md-paste-amd64" "$ROOT/cmd/md-paste"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$DIST/md-paste-arm64" "$ROOT/cmd/md-paste"
lipo -create -output "$DIST/md-paste-bin" "$DIST/md-paste-amd64" "$DIST/md-paste-arm64"
rm -f "$DIST/md-paste-amd64" "$DIST/md-paste-arm64"

echo "Assembling $APP..."
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
mv "$DIST/md-paste-bin" "$APP/Contents/MacOS/md-paste-bin"

# Wrapper so double-clicking the .app starts menu bar mode. The same binary
# still works as a plain CLI when invoked directly (e.g. the login item runs
# `md-paste-bin menubar`).
cat >"$APP/Contents/MacOS/md-paste" <<'WRAP'
#!/bin/sh
DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$DIR/md-paste-bin" menubar "$@"
WRAP
chmod +x "$APP/Contents/MacOS/md-paste" "$APP/Contents/MacOS/md-paste-bin"

sed "s/__VERSION__/${VERSION}/g" "$ROOT/build/Info.plist.tmpl" >"$APP/Contents/Info.plist"

if [ ! -f "$ROOT/build/icon.icns" ]; then
	"$ROOT/scripts/make-icns.sh"
fi
cp "$ROOT/build/icon.icns" "$APP/Contents/Resources/icon.icns"

echo "Ad-hoc signing..."
codesign --force --deep --sign - "$APP"

echo "Zipping..."
ZIP="$DIST/md-paste_${VERSION}_macos.zip"
rm -f "$ZIP"
ditto -c -k --keepParent "$APP" "$ZIP"
echo "Created $ZIP"
