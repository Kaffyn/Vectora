.PHONY: setup dev build build-tray build-cli build-web

# Compila o projeto base escondendo a janela de terminal no Windows (windowsgui)
build:
	go build -ldflags "-H=windowsgui -w -s" -o vectora.exe ./cmd/vectora

dev:
	go build -o vectora.exe ./cmd/vectora
	./vectora.exe
