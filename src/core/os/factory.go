package os

import (
	"runtime"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain"
	oslinux "github.com/Kaffyn/Vectora/src/os/linux"
	osmacos "github.com/Kaffyn/Vectora/src/os/macos"
	oswindows "github.com/Kaffyn/Vectora/src/os/windows"
)

// NewOSManager é a factory que retorna a implementação correta de OSManager baseada no runtime.
func NewOSManager(paths config.VectoraPaths) domain.OSManager {
	switch runtime.GOOS {
	case "windows":
		return oswindows.NewWindowsManager(paths)
	case "linux":
		return oslinux.NewLinuxManager(paths)
	case "darwin": // macOS
		return osmacos.NewMacOSManager(paths)
	default:
		// Fallback para linux caso desconhecido, ou panics se preferir
		return oslinux.NewLinuxManager(paths)
	}
}
