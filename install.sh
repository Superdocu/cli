#!/bin/sh
# Install the latest superdocu CLI release for the current OS/arch.
#
#   curl -fsSL https://raw.githubusercontent.com/Superdocu/cli/main/install.sh | sh
#
# Override the version with VERSION=v1.2.3 and the target dir with BINDIR=...
set -eu

REPO="Superdocu/cli"
BIN="superdocu"
BINDIR="${BINDIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
	linux | darwin) ;;
	*) echo "unsupported OS: $os — download the Windows zip from the Releases page" >&2; exit 1 ;;
esac

version="${VERSION:-}"
if [ -z "$version" ]; then
	version=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest" | sed 's#.*/tag/##')
fi
[ -n "$version" ] || { echo "could not resolve the latest version" >&2; exit 1; }

num=${version#v}
tarball="${BIN}_${num}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$version/$tarball"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $url"
curl -fsSL "$url" -o "$tmp/$tarball"
tar -xzf "$tmp/$tarball" -C "$tmp"

if [ -w "$BINDIR" ]; then
	install -m 0755 "$tmp/$BIN" "$BINDIR/$BIN"
else
	echo "Installing to $BINDIR (needs sudo)"
	sudo install -m 0755 "$tmp/$BIN" "$BINDIR/$BIN"
fi

echo "Installed $BIN $version to $BINDIR/$BIN"
