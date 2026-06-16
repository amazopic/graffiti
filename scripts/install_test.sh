#!/bin/sh
# Offline test for install.sh: detect_target (mocked uname) + verify_checksum.
set -eu
cd "$(dirname "$0")/.."
GRAFFITI_INSTALL_LIB=1 . ./scripts/install.sh

fail=0
expect() { # expect DESC EXPECTED ACTUAL
	if [ "$2" = "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want [$2] got [$3]"; fail=1; fi
}

# detect_target with mocked uname.
check_target() { # MOCK_OS MOCK_ARCH EXPECTED
	uname() { case "$1" in -s) echo "$_OS" ;; -m) echo "$_ARCH" ;; esac; }
	_OS=$1 _ARCH=$2
	got=$(detect_target) || got="ERR"
	expect "detect $1/$2" "$3" "$got"
}
check_target Darwin arm64           graffiti-darwin-arm64
check_target Darwin x86_64          graffiti-darwin-amd64
check_target Linux  x86_64          graffiti-linux-amd64
check_target Linux  aarch64         graffiti-linux-arm64
check_target MINGW64_NT-10.0 x86_64 graffiti-windows-amd64.exe

# unsupported platform → error (non-zero), reported as ERR
uname() { case "$1" in -s) echo SunOS ;; -m) echo sparc ;; esac; }
got=$(detect_target 2>/dev/null) || got="ERR"
expect "detect unsupported" "ERR" "$got"
unset -f uname

# verify_checksum: build a manifest for a temp file and confirm match + tamper-fail.
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
printf 'hello graffiti' > "$tmp/graffiti-linux-amd64"
sum=$( (command -v sha256sum >/dev/null 2>&1 && sha256sum "$tmp/graffiti-linux-amd64" || shasum -a 256 "$tmp/graffiti-linux-amd64") | awk '{print $1}')
printf '%s  graffiti-linux-amd64\n' "$sum" > "$tmp/SHA256SUMS"
if verify_checksum "$tmp/graffiti-linux-amd64" "$tmp/SHA256SUMS" graffiti-linux-amd64; then echo "ok   - checksum match"; else echo "FAIL - checksum match"; fail=1; fi
printf 'tampered' > "$tmp/graffiti-linux-amd64"
if verify_checksum "$tmp/graffiti-linux-amd64" "$tmp/SHA256SUMS" graffiti-linux-amd64 2>/dev/null; then echo "FAIL - tamper should fail"; fail=1; else echo "ok   - tamper detected"; fi

[ "$fail" = 0 ] && echo "ALL INSTALL TESTS PASSED" || { echo "INSTALL TESTS FAILED"; exit 1; }
