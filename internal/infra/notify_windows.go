//go:build windows

package infra

import (
	"log"
	"os"

	"github.com/go-toast/toast"
)

func NotifyOS(title, message string) error {
	// Dica de Windows 11 UX:
	// O campo "toast.Notification.Icon" desenha a imagem no CORPO GIGANTE da notificação (AppLogoOverride).
	// Para forçar o ícone a ir para a parte esquerda superior (O Header do App), o AppID
	// deve apontar diretamente pro binário físico `.exe` que contém o manifesto de ícone injetado pelo rsrc.
	execPath, err := os.Executable()
	if err != nil {
		execPath = "Vectora"
	}

	notification := toast.Notification{
		AppID:   execPath, 
		Title:   title,
		Message: message,
		// Removido 'Icon' explícito para não flutuar o icone gigante fora do lugar.
	}
	
	err = notification.Push()
	if err != nil {
		log.Println("Win32 Toast Notification falhou na infraestrutura:", err)
	}
	return err
}
