#!/usr/bin/env bash
# Publish the md-paste Homebrew Cask for the .app bundle to the tap repo.
# Usage: scripts/publish-cask.sh <version> <zip-path>
# Requires the TAP_GITHUB_TOKEN environment variable.
set -euo pipefail

VERSION="${1:?usage: publish-cask.sh <version> <zip-path>}"
ZIP="${2:?usage: publish-cask.sh <version> <zip-path>}"
: "${TAP_GITHUB_TOKEN:?TAP_GITHUB_TOKEN is required}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SHA="$(shasum -a 256 "$ZIP" | awk '{print $1}')"

TMP="$(mktemp -d)"
git clone --depth 1 "https://x-access-token:${TAP_GITHUB_TOKEN}@github.com/stn1slv/homebrew-tap.git" "$TMP"

mkdir -p "$TMP/Casks"
sed -e "s/__VERSION__/${VERSION}/g" -e "s/__SHA256__/${SHA}/g" \
	"$ROOT/build/md-paste.rb.tmpl" >"$TMP/Casks/md-paste.rb"

git -C "$TMP" add Casks/md-paste.rb
if git -C "$TMP" diff --cached --quiet; then
	echo "Cask already up to date; nothing to publish."
	exit 0
fi
git -C "$TMP" \
	-c user.name="github-actions[bot]" \
	-c user.email="github-actions[bot]@users.noreply.github.com" \
	commit -m "chore: update md-paste cask to ${VERSION}"
git -C "$TMP" push origin HEAD
echo "Published md-paste cask ${VERSION}"
