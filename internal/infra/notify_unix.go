//go:build !windows

package infra

import (
	"github.com/gen2brain/beeep"
)

func NotifyOS(title, message string) error {
	// Mac e Linux lidam graciosamente com Beeep com Application Name Inference via BIN.
	return beeep.Notify(title, message, "")
}
