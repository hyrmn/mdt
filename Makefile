GO ?= go
BIN_DIR ?= bin
DIST_DIR ?= dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf 'dev')
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
GOFLAGS ?=
LDFLAGS ?= -s -w
RELEASE_PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

export GOOS GOARCH

EXE := $(if $(filter windows,$(GOOS)),.exe,)

.DEFAULT_GOAL := all
.PHONY: all build mdv mde install test race vet fmt fmt-check tidy check release clean help

all: build

## Build both utilities into BIN_DIR.
build: mdv mde

## Build only mdv.
mdv:
	@mkdir -p "$(BIN_DIR)"
	$(GO) build $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" -o "$(BIN_DIR)/mdv$(EXE)" ./cmd/mdv

## Build only mde.
mde:
	@mkdir -p "$(BIN_DIR)"
	$(GO) build $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" -o "$(BIN_DIR)/mde$(EXE)" ./cmd/mde

## Install both utilities using go install.
install:
	$(GO) install $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" ./cmd/mdv ./cmd/mde

## Run all unit and simulated-TUI tests.
test:
	$(GO) test ./...

## Run tests with the race detector.
race:
	$(GO) test -race ./...

## Run go vet.
vet:
	$(GO) vet ./...

## Format all Go source files.
fmt:
	gofmt -w cmd internal

## Fail if any Go source file is not formatted.
fmt-check:
	@files="$$(gofmt -l cmd internal)"; \
	if [ -n "$$files" ]; then \
		printf 'The following files need gofmt:\n%s\n' "$$files"; \
		exit 1; \
	fi

## Synchronize module dependencies.
tidy:
	$(GO) mod tidy

## Run formatting checks, vet, tests, and a native build.
check: fmt-check vet test build

## Build release archives for all RELEASE_PLATFORMS (requires tar, zip, and sha256sum).
release: check
	@rm -rf "$(DIST_DIR)"
	@mkdir -p "$(DIST_DIR)"
	@set -e; \
	for platform in $(RELEASE_PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		name="mdt_$(VERSION)_$${os}_$${arch}"; \
		staging="$(DIST_DIR)/$${name}"; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		mkdir -p "$$staging"; \
		for binary in mdv mde; do \
			printf 'Building %s for %s/%s\n' "$$binary" "$$os" "$$arch"; \
			CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" $(GO) build $(GOFLAGS) -trimpath \
				-ldflags "$(LDFLAGS)" -o "$$staging/$$binary$$ext" "./cmd/$$binary"; \
		done; \
		cp LICENSE README.md "$$staging/"; \
		if [ "$$os" = "windows" ]; then \
			(cd "$$staging" && zip -q -r "$(abspath $(DIST_DIR))/$$name.zip" .); \
		else \
			tar -czf "$(abspath $(DIST_DIR))/$$name.tar.gz" -C "$$staging" .; \
		fi; \
		rm -rf "$$staging"; \
	done
	@cd "$(DIST_DIR)" && for archive in *.tar.gz *.zip; do \
		if [ -f "$$archive" ]; then sha256sum "$$archive"; fi; \
	done > SHA256SUMS
	@printf 'Release archives and checksums are in %s\n' "$(DIST_DIR)"

## Remove local build and release output.
clean:
	rm -rf "$(BIN_DIR)" "$(DIST_DIR)"

## List available targets.
help:
	@printf '%s\n' \
		'make build       Build mdv and mde (default)' \
		'make mdv         Build only mdv' \
		'make mde         Build only mde' \
		'make install     Install both commands with go install' \
		'make test        Run all tests' \
		'make race        Run tests with the race detector' \
		'make vet         Run go vet' \
		'make fmt         Format Go source' \
		'make fmt-check   Check Go formatting' \
		'make tidy        Run go mod tidy' \
		'make check       Run formatting, vet, tests, and native build' \
		'make release     Build cross-platform archives and SHA256SUMS' \
		'make clean       Remove bin and dist output'
