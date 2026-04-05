.PHONY: setup dev build-all build-daemon build-app build-cli test clean help

help:
	@echo "Vectora Build System"
	@echo "===================="
	@echo "setup          - Setup dependências (Go + Node)"
	@echo "dev            - Run em modo desenvolvimento"
	@echo "build-daemon   - Compilar daemon"
	@echo "build-app      - Compilar Web UI (Wails)"
	@echo "build-all      - Compilar tudo"
	@echo "test           - Rodar testes"
	@echo "clean          - Limpar builds"

setup:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Node dependencies..."
	cd internal/app && npm install || bun install
	@echo "Setup completo!"

build-daemon:
	@echo "Building daemon..."
	mkdir -p build
	go build -o build/vectora ./cmd/vectora
	@echo "✓ Daemon built: build/vectora"

build-app:
	@echo "Building Web UI..."
	cd internal/app && npm run build || bun run build
	go build -o build/vectora-app ./cmd/vectora-app
	@echo "✓ App built: build/vectora-app"

build-cli:
	@echo "Building CLI..."
	mkdir -p build
	go build -o build/vectora-cli ./cmd/vectora
	@echo "✓ CLI built: build/vectora-cli"

build-all: build-daemon build-app build-cli
	@echo "✓ All builds complete!"

dev-daemon:
	@echo "Running daemon (dev mode)..."
	go run ./cmd/vectora/main.go

dev-app:
	@echo "Running Web UI (dev mode)..."
	cd internal/app && npm run dev || bun run dev

test:
	@echo "Running tests..."
	go test -v -race ./...
	go test -cover ./...

clean:
	@echo "Cleaning builds..."
	rm -rf build/
	rm -rf internal/app/.next
	rm -rf internal/app/out
	rm -rf internal/app/node_modules
	go clean -cache
	@echo "✓ Clean complete!"

.PHONY: all
all: setup build-all test
