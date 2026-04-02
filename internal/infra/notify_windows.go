//go:build windows

package infra

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Kaffyn/vectora/assets"
	"github.com/go-toast/toast"
)

func NotifyOS(title, message string) error {
	// A API do Action Center do Windows exige o caminho físico em Disco (Absolute Path) para desenhar imagens.
	// Como a nossa logo está imbutida na memória (Go Embed), nós extraimos pra folder temporário da CPU:
	iconPath := filepath.Join(os.TempDir(), "vectora_logo.ico")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		os.WriteFile(iconPath, assets.IconData, 0644)
	}

	notification := toast.Notification{
		AppID:   "Vectora", // Removido texto (Kaffyn), adotado nominal base
		Title:   title,
		Message: message,
		Icon:    iconPath,
	}
	
	err := notification.Push()
	if err != nil {
		log.Println("Win32 Toast Notification falhou na infraestrutura:", err)
	}
	return err
}
