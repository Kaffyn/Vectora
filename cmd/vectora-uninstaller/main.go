package main

import (
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
	"golang.org/x/sys/windows/registry"
)

func main() {
	beeep.Notify("Vectora", "Desinstalando o Vectora System App de forma limpa...", "")
	
	// Limpeza do Registro
	registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
	
	// Limpeza de Processos Mestre
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	
	// Exclui a própria espinha dorsal do Daemon do diretório se existir
	vectoraDaemonPath := filepath.Join(dir, "vectora.exe")
	if _, err := os.Stat(vectoraDaemonPath); err == nil {
		os.Remove(vectoraDaemonPath)
	}

	beeep.Notify("Vectora", "Software removido com sucesso do painel de controle.", "")
}
