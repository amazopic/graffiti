# graffiti build helpers. The grammar_subset tags ship only the Go grammar,
# keeping the binary small (~8MB) and CGO-free. Without them the code still
# builds, but links the full grammar set (~31MB).
TAGS := grammar_subset grammar_subset_go grammar_subset_gomod \
        grammar_subset_python grammar_subset_javascript grammar_subset_typescript \
        grammar_subset_rust grammar_subset_java grammar_subset_php
PKG  := ./cmd/graffiti

# VERSION is derived from git (tag/commit); release builds inject it via ldflags.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS         := -X main.version=$(VERSION)
RELEASE_LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet xcompile size-guard release

build:
	CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o graffiti $(PKG)

test:
	go test -tags "$(TAGS)" ./...

vet:
	go vet -tags "$(TAGS)" ./...

# Cross-compile the static binary for all v1 targets (spec §10).
xcompile:
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-darwin-arm64 $(PKG)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-darwin-amd64 $(PKG)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-linux-amd64  $(PKG)
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-linux-arm64  $(PKG)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$(RELEASE_LDFLAGS)" -o dist/graffiti-windows-amd64.exe $(PKG)
	$(MAKE) size-guard

# Binary size guard: fails the build if any dist/ binary exceeds 16MB (~16000000 bytes).
# A missing-build-tag regression balloons the binary from ~9MB to ~31MB; this catches it.
SIZE_LIMIT := 16000000
size-guard:
	@for f in dist/graffiti-darwin-arm64 dist/graffiti-darwin-amd64 dist/graffiti-linux-amd64 dist/graffiti-linux-arm64 dist/graffiti-windows-amd64.exe; do \
		size=$$(wc -c < "$$f"); \
		echo "size-guard: $$f = $$size bytes (limit $(SIZE_LIMIT))"; \
		if [ "$$size" -ge "$(SIZE_LIMIT)" ]; then \
			echo "ERROR: $$f exceeds size limit ($$size >= $(SIZE_LIMIT)) — subset build tags likely missing"; \
			exit 1; \
		fi; \
	done
	@echo "size-guard: all binaries within limit OK"

# release cross-compiles all targets (with size-guard via xcompile) and writes a
# SHA256SUMS manifest over the dist/ binaries. Used by .github/workflows/release.yml.
release: xcompile
	@cd dist && { \
		if command -v sha256sum >/dev/null 2>&1; then sha256sum graffiti-* > SHA256SUMS; \
		else shasum -a 256 graffiti-* > SHA256SUMS; fi; }
	@echo "release: wrote dist/SHA256SUMS"
	@cat dist/SHA256SUMS
