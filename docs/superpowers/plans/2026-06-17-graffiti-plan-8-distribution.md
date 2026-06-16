# graffiti Plan 8 — Distribution (one-command install)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make graffiti installable on a clean machine with one command — `curl -fsSL …/install.sh | sh` — backed by checksummed GitHub Releases that CI cross-compiles on every version tag, plus a `graffiti version` command (spec §10 distribution).

**Architecture:** Five static, CGO-free binaries (darwin/linux/windows × amd64/arm64; windows amd64 only in v1 — the existing `make xcompile` matrix) published as GitHub Releases with a `SHA256SUMS` manifest. A POSIX `scripts/install.sh` detects OS/arch, downloads the matching asset + `SHA256SUMS` from the latest (or a pinned) release, verifies the checksum, and installs to a writable bin dir. The binary's version is injected at build time via `-ldflags "-X main.version=…"` (from `git describe`). Two GitHub Actions workflows: `ci.yml` (vet/test/build on push & PR) and `release.yml` (on a `v*` tag: `make release` → upload binaries + `SHA256SUMS`).

**Tech Stack:** Go 1.26 (ldflags), POSIX `sh`, GNU make, GitHub Actions. No new Go dependencies; no runtime dependencies in the binary.

**What this plan does NOT do (deferred, with rationale):**
- **Upstream tree-sitter parity-diff CI** (the Plan-6 follow-up): it needs the `tree-sitter` CLI + npm + network to develop and verify iteratively, which this offline environment can't provide, so authoring it blind would be unverifiable. It stays deferred to a dedicated session with that toolchain.
- Homebrew tap / scoop / winget (spec §10 "later").

**Validation reality (read this):** the `version` command, the Makefile `release` target, and `install.sh`'s OS/arch detection + checksum verification are **fully runnable and tested here**. The network download path (`install.sh` fetching from GitHub) and the **GitHub Actions workflows cannot be executed offline** — they are validated *by construction* (mirroring the byte-for-byte commands of the proven `Makefile`) and reviewed structurally; their first live exercise is the first real push/tag after the repo is on GitHub. Each such step says so explicitly.

## File structure

```
cmd/graffiti/main.go        + version package var + `version`/`--version` command + usage line
cmd/graffiti/main_test.go   + version command test
Makefile                    + VERSION, -ldflags on build/xcompile, `release` target
scripts/install.sh          POSIX installer (sourceable functions + guarded main)        [new]
scripts/install_test.sh     offline shell test: mocked-uname detection + checksum verify  [new]
.github/workflows/ci.yml    push/PR: vet + test (both tag configs) + build + xcompile     [new]
.github/workflows/release.yml  v* tag: make release → GitHub Release w/ binaries + sums   [new]
README.md                   Install section
docs/superpowers/specs/2026-06-14-graffiti-design.md   §10 distribution-implemented note
```

The public release location is `github.com/evgeniy-achin/graffiti` (the module path). The repo currently has no remote; the installer/workflows go live once it is pushed to GitHub.

---

## Task 1: `graffiti version` command

**Files:**
- Modify: `cmd/graffiti/main.go`, `cmd/graffiti/main_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/graffiti/main_test.go`:

```go
func TestRun_Version(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		var out, errOut bytes.Buffer
		code := run([]string{"graffiti", arg}, bytes.NewReader(nil), &out, &errOut)
		if code != 0 {
			t.Fatalf("%s exit=%d stderr=%q", arg, code, errOut.String())
		}
		if !strings.Contains(out.String(), "graffiti ") {
			t.Fatalf("%s: expected 'graffiti <version>', got %q", arg, out.String())
		}
		if !strings.Contains(out.String(), version) {
			t.Fatalf("%s: output %q missing version %q", arg, out.String(), version)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./cmd/graffiti/ -run TestRun_Version -v`
Expected: FAIL — `version` identifier undefined (and the commands are unknown).

- [ ] **Step 3: Add the version var and command**

In `cmd/graffiti/main.go`, add a package-level var near the top (after the imports), with a doc comment:

```go
// version is the build version, injected at release time via
// -ldflags "-X main.version=<tag>". Defaults to "dev" for local builds.
var version = "dev"
```

Add cases to the `switch cmd` block (place them before `default:`):

```go
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, "graffiti "+version)
		return 0
```

- [ ] **Step 4: Add a usage line**

In `func usage`, add after the `init`/`link` lines:

```go
	fmt.Fprintln(w, "  version           print the graffiti version")
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./cmd/graffiti/ -run TestRun_Version -v`
Expected: PASS (prints "graffiti dev" in tests).

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat(cli): graffiti version command (ldflags-injected build version)"
```

---

## Task 2: Makefile version ldflags + `release` target

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add VERSION + LDFLAGS and thread them into build/xcompile**

At the top of `Makefile`, after the `PKG :=` line, add:

```makefile
# VERSION is derived from git (tag/commit); release builds inject it via ldflags.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS        := -X main.version=$(VERSION)
RELEASE_LDFLAGS := -s -w -X main.version=$(VERSION)
```

Update `.PHONY` to include `release`:

```makefile
.PHONY: build test vet xcompile size-guard release
```

Change the `build` recipe to inject the version:

```makefile
build:
	CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o graffiti $(PKG)
```

Change each `xcompile` line to add `-ldflags "$(RELEASE_LDFLAGS)"`. For example the first line becomes:

```makefile
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-darwin-arm64 $(PKG)
```

Apply the same `-ldflags "$(RELEASE_LDFLAGS)"` insertion to all five `xcompile` build lines (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64, windows-amd64.exe).

- [ ] **Step 2: Add the `release` target**

Append to `Makefile`:

```makefile
# release cross-compiles all targets (with size-guard via xcompile) and writes a
# SHA256SUMS manifest over the dist/ binaries. Used by .github/workflows/release.yml.
release: xcompile
	@cd dist && { \
		if command -v sha256sum >/dev/null 2>&1; then sha256sum graffiti-* > SHA256SUMS; \
		else shasum -a 256 graffiti-* > SHA256SUMS; fi; }
	@echo "release: wrote dist/SHA256SUMS"
	@cat dist/SHA256SUMS
```

- [ ] **Step 3: Validate the release target (runnable here)**

Run: `make release`
Expected: 5 binaries built under `dist/`, size-guard prints "all binaries within limit OK", then `dist/SHA256SUMS` is written and printed with **5 lines** (one per binary).

Verify count: `test "$(grep -c graffiti- dist/SHA256SUMS)" = 5 && echo OK`
Expected: `OK`

- [ ] **Step 4: Validate ldflags version injection (runnable here)**

```bash
go build -tags "grammar_subset grammar_subset_go grammar_subset_gomod" -ldflags "-X main.version=v9.9.9-test" -o /tmp/gver ./cmd/graffiti
/tmp/gver version
rm -f /tmp/gver
```
Expected: prints `graffiti v9.9.9-test` (proves `-X main.version` injection works).

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "build(make): inject version via ldflags; add release target + SHA256SUMS"
```

---

## Task 3: `scripts/install.sh` + offline shell tests

**Files:**
- Create: `scripts/install.sh`, `scripts/install_test.sh`

- [ ] **Step 1: Write `scripts/install.sh`**

A POSIX `sh` script whose core logic lives in sourceable functions (so the test can call them without running `main`). The asset names match the `make xcompile` outputs exactly.

```sh
#!/bin/sh
# install.sh — install the graffiti binary from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/evgeniy-achin/graffiti/main/scripts/install.sh | sh
#
# Environment:
#   GRAFFITI_VERSION  release tag to install (default: latest)
#   INSTALL_DIR       install directory (default: /usr/local/bin, fallback ~/.local/bin)
set -eu

REPO="evgeniy-achin/graffiti"
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
```

- [ ] **Step 2: Write `scripts/install_test.sh`**

A POSIX test that sources the library, mocks `uname`, and checks detection + checksum verification (no network):

```sh
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
check_target Darwin arm64  graffiti-darwin-arm64
check_target Darwin x86_64 graffiti-darwin-amd64
check_target Linux  x86_64 graffiti-linux-amd64
check_target Linux  aarch64 graffiti-linux-arm64
check_target MINGW64_NT-10.0 x86_64 graffiti-windows-amd64.exe

# unsupported platform → error (non-zero), printed as ERR
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
```

- [ ] **Step 3: Syntax-check and run the tests (runnable here)**

```bash
sh -n scripts/install.sh
chmod +x scripts/install.sh scripts/install_test.sh
sh scripts/install_test.sh
```
Expected: `sh -n` is silent (valid syntax); the test prints all `ok` lines ending in `ALL INSTALL TESTS PASSED`.

NOTE: the network download path in `main()` is exercised only against a real GitHub Release (not offline); detection + checksum verification — the parts most likely to be wrong — are fully covered above.

- [ ] **Step 4: Commit**

```bash
git add scripts/install.sh scripts/install_test.sh
git commit -m "feat(install): POSIX install.sh (detect/verify/download) + offline shell tests"
```

---

## Task 4: GitHub Actions — `ci.yml` + `release.yml`

**Files:**
- Create: `.github/workflows/ci.yml`, `.github/workflows/release.yml`

These cannot be executed offline; each step mirrors a command from the proven `Makefile`, so they are validated by construction. They activate when the repo is pushed to GitHub.

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:

jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: vet
        run: make vet
      - name: test (subset build tags)
        run: make test
      - name: test (default tags, full grammar embed)
        run: go test ./...
      - name: build + cross-compile + size guard
        run: make xcompile
```

- [ ] **Step 2: Write `.github/workflows/release.yml`**

```yaml
name: Release
on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # full history so `git describe --tags` resolves the tag
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: cross-compile + checksums
        run: make release
      - name: publish GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/graffiti-darwin-arm64
            dist/graffiti-darwin-amd64
            dist/graffiti-linux-amd64
            dist/graffiti-linux-arm64
            dist/graffiti-windows-amd64.exe
            dist/SHA256SUMS
          generate_release_notes: true
```

- [ ] **Step 3: Structural sanity check (runnable here)**

```bash
ls .github/workflows/ci.yml .github/workflows/release.yml
# basic YAML well-formedness via python (no actionlint available):
python3 -c "import yaml,sys; [yaml.safe_load(open(f)) for f in sys.argv[1:]]; print('yaml OK')" .github/workflows/ci.yml .github/workflows/release.yml
```
Expected: both files exist; `yaml OK`. (Full workflow validation happens on the first push/tag to GitHub.)

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yml .github/workflows/release.yml
git commit -m "ci: GitHub Actions for test matrix and tagged releases"
```

---

## Task 5: docs + full verification

**Files:**
- Modify: `README.md`, `docs/superpowers/specs/2026-06-14-graffiti-design.md`

- [ ] **Step 1: README — add an Install section** (near the top, before/around `## Build`):

```markdown
## Install

```bash
curl -fsSL https://raw.githubusercontent.com/evgeniy-achin/graffiti/main/scripts/install.sh | sh
```

Pin a version or directory:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/evgeniy-achin/graffiti/main/scripts/install.sh)"
```

The installer picks the right static binary for your OS/arch, verifies its SHA256
against the release manifest, and installs it. Verify with `graffiti version`.
Or build from source: `make build` (see below).
```

- [ ] **Step 2: Spec §10 — append a distribution-implemented note**

Add after the §10 Distribution bullet:

```markdown
**Implemented (Plan 8, 2026-06-17):** `make release` cross-compiles the five static CGO-free targets (darwin/linux ×amd64/arm64 + windows/amd64) with the version injected via `-ldflags "-X main.version=$(git describe)"`, then writes a `SHA256SUMS` manifest. `.github/workflows/release.yml` runs `make release` on a `v*` tag and publishes the binaries + manifest as a GitHub Release; `ci.yml` runs vet + tests (both tag configs) + cross-compile on push/PR. `scripts/install.sh` (POSIX, curl|sh) detects OS/arch, downloads the matching asset + `SHA256SUMS` from the latest/pinned release, verifies the checksum, and installs to a writable bin dir; `graffiti version` reports the build. Homebrew/scoop/winget and the upstream-tree-sitter parity-diff CI remain deferred.
```

- [ ] **Step 3: Full verification (runnable here)**

```bash
make vet
make test
go test ./...
go mod tidy && git diff --exit-code go.mod go.sum
make release
test "$(grep -c graffiti- dist/SHA256SUMS)" = 5 && echo "SHA256SUMS OK"
sh -n scripts/install.sh && sh scripts/install_test.sh
go build -tags "grammar_subset grammar_subset_go grammar_subset_gomod" -ldflags "-X main.version=v0.0.0-verify" -o /tmp/gv ./cmd/graffiti && /tmp/gv version && rm -f /tmp/gv
```
Expected: vet clean; both test configs green; zero new deps; `make release` builds 5 binaries under the size guard and a 5-line `SHA256SUMS`; install tests pass; the ldflags build prints `graffiti v0.0.0-verify`.

- [ ] **Step 4: Commit**

```bash
git add README.md docs/superpowers/specs/2026-06-14-graffiti-design.md
git commit -m "docs: install instructions + spec §10 distribution note"
```

---

## Self-review checklist (run before merge)

1. **Spec §10 coverage:** single-command install ✓ (install.sh); cross-compiled static binaries ✓ (make release/xcompile, unchanged matrix); GitHub Releases via CI ✓ (release.yml); `curl … | sh` ✓; version reporting ✓. Homebrew/scoop + parity-diff explicitly deferred.
2. **Asset-name consistency:** `detect_target` output, the `xcompile` `-o` names, and `release.yml`'s `files:` list are byte-identical (`graffiti-<os>-<arch>[.exe]`). The install test asserts the five names.
3. **Determinism / no regressions:** ldflags only adds a version string + (release) strips symbols; no behavior change. Both test configs stay green; size guard still holds (stripped binaries are smaller).
4. **No new deps:** stdlib only in Go; `go mod tidy` no-op. CI uses standard published actions (checkout/setup-go/action-gh-release) — not part of the binary.
5. **Honest validation:** every offline-unverifiable step (network download, GitHub Actions) is labeled as validated-by-construction; the verifiable core (version, detection, checksum, release target) is actually run.

## Deferred follow-ups (record in memory, non-blocking)

- **Upstream tree-sitter parity-diff CI** (needs `tree-sitter` CLI + npm + network; develop in an online session).
- Homebrew tap, scoop/winget manifests (spec §10 "later").
- `install.sh` GPG/sigstore signature verification (currently SHA256 manifest only — sufficient for tamper detection over HTTPS, not provenance).
- Reproducible-build flags (`-trimpath`, `-buildvcs`) for byte-identical release binaries across machines.
