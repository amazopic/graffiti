# graffiti build helpers. The grammar_subset tags ship only the Go grammar,
# keeping the binary small (~8MB) and CGO-free. Without them the code still
# builds, but links the full grammar set (~31MB).
TAGS := grammar_subset grammar_subset_go grammar_subset_gomod
PKG  := ./cmd/graffiti

.PHONY: build test vet xcompile

build:
	CGO_ENABLED=0 go build -tags "$(TAGS)" -o graffiti $(PKG)

test:
	go test -tags "$(TAGS)" ./...

vet:
	go vet -tags "$(TAGS)" ./...

# Cross-compile the static binary for all v1 targets (spec §10).
xcompile:
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-darwin-arm64 $(PKG)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-darwin-amd64 $(PKG)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-linux-amd64  $(PKG)
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-linux-arm64  $(PKG)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-windows-amd64.exe $(PKG)
