.PHONY: help run generate test lint build deploy docker-build docker-run clean

# Variáveis
GO := go
PROTOC := protoc
GRPC_PORT ?= 3000
BIN_NAME := vectora-index

help:
	@echo "Vectora Index Service - Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make run              - Rodar servidor localmente"
	@echo "  make generate         - Gerar código protobuf"
	@echo "  make test             - Rodar testes"
	@echo "  make lint             - Verificar código (golangci-lint)"
	@echo "  make fmt              - Formatar código"
	@echo "  make build            - Build binário"
	@echo "  make deploy           - Deploy em Vercel"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Rodar em Docker"
	@echo "  make clean            - Limpar arquivos gerados"
	@echo "  make deps             - Baixar dependências"
	@echo "  make grpcurl-install  - Instalar grpcurl"
	@echo ""

deps:
	@echo "[INFO] Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

generate:
	@echo "[INFO] Generating protobuf code..."
	@mkdir -p api/gen/go/vectora/index/v1
	$(PROTOC) -I api/v1 \
		--go_out=api/gen/go \
		--go-grpc_out=api/gen/go \
		api/v1/index.proto
	@echo "[INFO] Protobuf generation complete!"

run: generate
	@echo "[INFO] Starting Vectora Index Service on localhost:$(GRPC_PORT)..."
	GRPC_PORT=$(GRPC_PORT) GRPC_HOST=0.0.0.0 $(GO) run cmd/server/main.go

test: generate
	@echo "[INFO] Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "[INFO] Coverage report:"
	$(GO) tool cover -func=coverage.out | tail -1

fmt:
	@echo "[INFO] Formatting code..."
	$(GO) fmt ./...
	@echo "[INFO] Format complete!"

lint:
	@echo "[INFO] Linting code..."
	golangci-lint run ./...

build: generate
	@echo "[INFO] Building binary..."
	$(GO) build -ldflags="-s -w" \
		-o bin/$(BIN_NAME) \
		./cmd/server
	@echo "[INFO] Build complete: bin/$(BIN_NAME)"

deploy:
	@echo "[INFO] Deploying to Vercel..."
	vercel deploy --prod

docker-build: build
	@echo "[INFO] Building Docker image..."
	docker build -t $(BIN_NAME):latest .
	@echo "[INFO] Docker build complete!"

docker-run: docker-build
	@echo "[INFO] Running Docker container..."
	docker run -p $(GRPC_PORT):3000 \
		--env-file .env \
		$(BIN_NAME):latest

grpcurl-install:
	@echo "[INFO] Installing grpcurl..."
	$(GO) install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@echo "[INFO] grpcurl installed!"

grpc-health-check: grpcurl-install
	@echo "[INFO] Checking gRPC server health..."
	grpcurl -plaintext localhost:$(GRPC_PORT) vectora.index.v1.IndexService/Health

clean:
	@echo "[INFO] Cleaning generated files..."
	rm -rf api/gen
	rm -rf bin
	rm -f coverage.out
	$(GO) clean
	@echo "[INFO] Clean complete!"

install-tools:
	@echo "[INFO] Installing development tools..."
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@echo "[INFO] Tools installed!"

# Verificação de dependências
check-protoc:
	@command -v protoc >/dev/null 2>&1 || { echo "protoc is not installed. Install it first!"; exit 1; }
	@echo "[OK] protoc is installed"

check-go:
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed. Install it first!"; exit 1; }
	@echo "[OK] Go is installed: $(shell $(GO) version)"

all: check-go check-protoc deps generate build
	@echo "[OK] Build complete!"
