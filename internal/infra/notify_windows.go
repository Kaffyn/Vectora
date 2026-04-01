//go:build windows

package infra

import (
	"log"
	"os"

	"github.com/go-toast/toast"
)

func NotifyOS(title, message string) error {
	// Windows 11 UX Tip:
	// The "toast.Notification.Icon" field draws the image in the LARGE BODY of the notification (AppLogoOverride).
	// To force the icon to the top-left (App Header), the AppID
	// must point directly to the physical `.exe` binary that contains the icon manifest injected by rsrc.
	execPath, err := os.Executable()
	if err != nil {
		execPath = "Vectora"
	}

	notification := toast.Notification{
		AppID:   execPath, 
		Title:   title,
		Message: message,
		// Removed explicit 'Icon' to avoid floating a massive icon out of place.
	}
	
	err = notification.Push()
	if err != nil {
		log.Println("Win32 Toast Notification failed in infrastructure:", err)
	}
	return err
}
