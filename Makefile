# skim — build the binary into ./dist
#
#   make            build the frontend, then the native binary (dist/skim)
#   make dist       build the frontend, then cross-compile for win/mac/linux
#   make web        build just the Svelte UI (web/dist)
#   make run ARGS="http://localhost:3000"
#   make clean

# Version stamped into the binary (-v); derived from git, overridable: make VERSION=x.y.z
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO_LDFLAGS := -s -w -X main.version=$(VERSION)

# Fully static binaries (pure-Go DNS resolver) so they run on any distro / box.
export CGO_ENABLED := 0

.PHONY: all web dev dist run clean verify-dist

all: web
	go build -ldflags "$(GO_LDFLAGS)" -o dist/skim .

# Build the Svelte frontend into web/dist (embedded into the Go binary).
web:
	cd web && pnpm install && pnpm build

# web/dist is committed (embedded via go:embed) so the module is `go install`-able
# without a Node toolchain. Rebuild and fail if it drifted from web/src — run this
# in CI / before tagging so a stale commit can't ship an old UI.
verify-dist: web
	git diff --exit-code -- web/dist

# Frontend dev with HMR. Vite serves the UI and auto-runs the Go backend,
# proxying /api to it — edit web/src/** and see changes live.
dev:
	cd web && pnpm install && pnpm dev

# Static, single-file binaries — ~7MB each, no runtime needed.
dist: web
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o dist/skim-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o dist/skim-linux-arm64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o dist/skim-macos-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o dist/skim-macos-amd64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o dist/skim-windows-amd64.exe .

run: web
	go run . $(ARGS)

clean:
	rm -rf dist web/dist web/node_modules
