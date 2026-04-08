//go:build windows

package infra

import (
	"log"
	"os"

	"github.com/go-toast/toast"
)

func NotifyOS(title, message string) error {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "Vectora"
	}

	notification := toast.Notification{
		AppID:   execPath,
		Title:   title,
		Message: message,
	}

	err = notification.Push()
	if err != nil {
		log.Println("Win32 Toast Notification failed in infrastructure:", err)
	}
	return err
}
