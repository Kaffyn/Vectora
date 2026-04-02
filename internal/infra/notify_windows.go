//go:build windows

package infra

import (
	"log"

	"github.com/go-toast/toast"
)

// NotifyOS dispara notificações de Desktop System native Action Center 
// usando injeção de AUMID customizado (AppID), erradicando a falha "DefaultAppName".
func NotifyOS(title, message string) error {
	notification := toast.Notification{
		AppID:   "Vectora (Kaffyn)", // Resolve The Registry App Name Mismatch
		Title:   title,
		Message: message,
	}
	err := notification.Push()
	if err != nil {
		log.Println("Win32 Toast Notification falhou na infraestrutura:", err)
	}
	return err
}
