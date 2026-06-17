#!/bin/sh
# install.sh — install the graffiti binary from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
#
# Environment:
#   GRAFFITI_VERSION  release tag to install (default: latest)
#   INSTALL_DIR       install directory (default: /usr/local/bin, fallback ~/.local/bin)
set -eu

REPO="amazopic/graffiti"
BIN="graffiti"

# detect_target prints the release asset name for the current OS/arch, e.g.
# "graffiti-darwin-arm64" or "graffiti-windows-amd64.exe". Returns non-zero on
# an unsupported platform.
detect_target() {
	os=$(uname -s)
	arch=$(uname -m)
	case "$os" in
		Linux) os=linux ;;
		Darwin) os=darwin ;;
		MINGW* | MSYS* | CYGWIN* | Windows_NT) os=windows ;;
		*) echo "graffiti: unsupported OS: $os" >&2; return 1 ;;
	esac
	case "$arch" in
		x86_64 | amd64) arch=amd64 ;;
		arm64 | aarch64) arch=arm64 ;;
		*) echo "graffiti: unsupported architecture: $arch" >&2; return 1 ;;
	esac
	if [ "$os" = windows ] && [ "$arch" != amd64 ]; then
		echo "graffiti: only windows/amd64 is published in v1" >&2
		return 1
	fi
	ext=""
	[ "$os" = windows ] && ext=".exe"
	printf '%s-%s-%s%s' "$BIN" "$os" "$arch" "$ext"
}

# sha256_of prints the lowercase-hex sha256 of a file using whichever tool exists.
sha256_of() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

# verify_checksum FILE SUMSFILE ASSET — confirms FILE's sha256 matches the line
# for ASSET in SUMSFILE (a `sha256  name` manifest). Returns non-zero on mismatch.
verify_checksum() {
	_file=$1; _sums=$2; _asset=$3
	_want=$(awk -v a="$_asset" '$2 == a || $2 == "*"a {print $1}' "$_sums" | head -n1)
	if [ -z "$_want" ]; then
		echo "graffiti: no checksum for $_asset in manifest" >&2
		return 1
	fi
	_got=$(sha256_of "$_file")
	if [ "$_want" != "$_got" ]; then
		echo "graffiti: checksum mismatch for $_asset (want $_want, got $_got)" >&2
		return 1
	fi
}

# fetch URL OUT — download URL to OUT using curl or wget.
fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		echo "graffiti: need curl or wget" >&2
		return 1
	fi
}

# choose_install_dir prints a writable bin directory.
choose_install_dir() {
	if [ -n "${INSTALL_DIR:-}" ]; then
		echo "$INSTALL_DIR"; return 0
	fi
	if [ -w /usr/local/bin ] 2>/dev/null; then
		echo /usr/local/bin; return 0
	fi
	echo "$HOME/.local/bin"
}

main() {
	target=$(detect_target)
	ver="${GRAFFITI_VERSION:-latest}"
	if [ "$ver" = latest ]; then
		base="https://github.com/$REPO/releases/latest/download"
	else
		base="https://github.com/$REPO/releases/download/$ver"
	fi

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	echo "graffiti: downloading $target ($ver)…"
	fetch "$base/$target" "$tmp/$target"
	fetch "$base/SHA256SUMS" "$tmp/SHA256SUMS"
	verify_checksum "$tmp/$target" "$tmp/SHA256SUMS" "$target"

	dir=$(choose_install_dir)
	mkdir -p "$dir"
	dest="$dir/$BIN"
	case "$target" in *.exe) dest="$dir/$BIN.exe" ;; esac
	mv "$tmp/$target" "$dest"
	chmod +x "$dest"

	echo "graffiti: installed to $dest"
	case ":$PATH:" in
		*":$dir:"*) ;;
		*) echo "graffiti: add $dir to your PATH (e.g. export PATH=\"$dir:\$PATH\")" ;;
	esac
	"$dest" version || true
}

# Run main unless sourced for testing (GRAFFITI_INSTALL_LIB=1 defines functions only).
if [ "${GRAFFITI_INSTALL_LIB:-0}" != 1 ]; then
	main "$@"
fi
