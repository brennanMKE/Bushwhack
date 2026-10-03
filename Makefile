# Bushwhack build. The Go binary embeds the built Svelte SPA (web/build), so
# build the web app first.

VERSION ?= $(shell git describe --always --dirty 2>/dev/null || date -u +%Y%m%d%H%M)
LDFLAGS := -s -w -X main.version=$(VERSION)
BIN     := bin

.PHONY: all web build linux test check dev demo run clean

all: build

web/node_modules: web/package-lock.json
	cd web && npm ci
	touch $@

web: web/node_modules
	cd web && npm run build

build: web
	go build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN)/bushwhack-server ./cmd/server
	go build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN)/bushwhack ./cmd/bushwhack

# photon is a t4g (Graviton, arm64) running Amazon Linux 2023.
linux: web
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN)/bushwhack-linux-arm64 ./cmd/server

test: web
	go vet ./...
	go test ./...
	cd web && npm test && npm run check

# Regenerate the hero animation from the Classic Jack pattern.
demo:
	go run ./cmd/gen-demo > web/src/lib/demo.json

run: build
	PORT=8080 $(BIN)/bushwhack-server

# Hot-reloading UI on :5173 proxying /api to a Go server on :8080.
dev:
	@echo "Run 'make run' in another terminal, then open http://localhost:5173"
	cd web && npm run dev

clean:
	rm -rf $(BIN) web/build
