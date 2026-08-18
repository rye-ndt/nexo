WAILS ?= $(shell command -v wails 2>/dev/null || echo $(HOME)/go/bin/wails)

.PHONY: dev dev-go build run mcp-shim test

# Start the app. Frontend hot-reloads; Go changes need a restart of this target.
dev:
	$(WAILS) dev -nogorebuild

# Same, but rebuilds and relaunches on Go changes. macOS steals focus each time.
dev-go:
	$(WAILS) dev

# Build the production .app bundle.
build:
	$(WAILS) build

# Build and launch the production app.
run: build
	./build/bin/Nexo.app/Contents/MacOS/nexo

mcp-shim:
	go build -o build/bin/nexo-mcp ./cmd/nexo-mcp

test:
	go test ./...
