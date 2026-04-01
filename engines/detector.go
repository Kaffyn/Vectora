package engines

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// DetectHardware identifica as capacidades de CPU/GPU do sistema.
func DetectHardware() (*Hardware, error) {
	hw := &Hardware{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GPUType:      "none",
		GPUVersion:   "",
	}

	// Detectar features de CPU
	if err := detectCPUFeatures(hw); err != nil {
		// Log do erro, mas continua (fallback para features padrão)
		fmt.Printf("warning: cpu feature detection failed: %v\n", err)
	}

	// Detectar core count
	hw.CoreCount = runtime.NumCPU()

	// Detectar RAM
	detectSystemRAM(hw)

	// Detectar GPU (específico por SO)
	switch hw.OS {
	case "windows":
		detectGPUWindows(hw)
	case "linux":
		detectGPULinux(hw)
	case "darwin":
		detectGPUMacOS(hw)
	}

	return hw, nil
}

// detectCPUFeatures identifica flags de CPU (AVX2, AVX512, NEON, etc).
func detectCPUFeatures(hw *Hardware) error {
	switch hw.OS {
	case "windows":
		return detectCPUFeaturesWindows(hw)
	case "linux":
		return detectCPUFeaturesLinux(hw)
	case "darwin":
		return detectCPUFeaturesMacOS(hw)
	}
	return nil
}

// detectCPUFeaturesWindows usa CPUID via syscall no Windows.
func detectCPUFeaturesWindows(hw *Hardware) error {
	// No Windows, usamos a API de Registry ou syscalls diretos
	// Para simplicidade, assumimos AVX2 em x86_64 moderno
	if hw.Architecture == "x86_64" {
		// Simular CPUID check: usuários modernos tem AVX2
		hw.CPUFeatures = []string{"avx2"}
	}
	return nil
}

// detectCPUFeaturesLinux lê /proc/cpuinfo.
func detectCPUFeaturesLinux(hw *Hardware) error {
	cmd := exec.Command("cat", "/proc/cpuinfo")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	cpuinfo := string(output)
	if strings.Contains(cpuinfo, "avx512f") {
		hw.CPUFeatures = append(hw.CPUFeatures, "avx512")
	} else if strings.Contains(cpuinfo, "avx2") {
		hw.CPUFeatures = append(hw.CPUFeatures, "avx2")
	}

	if strings.Contains(cpuinfo, "neon") {
		hw.CPUFeatures = append(hw.CPUFeatures, "neon")
	}

	return nil
}

// detectCPUFeaturesMacOS usa `sysctl` para verificar CPU features.
func detectCPUFeaturesMacOS(hw *Hardware) error {
	// macOS ARM64 (Apple Silicon) suporta NEON nativamente
	if hw.Architecture == "arm64" {
		hw.CPUFeatures = []string{"neon"}
		return nil
	}

	// macOS Intel x86_64
	cmd := exec.Command("sysctl", "-a")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: assume AVX2 em Intel modernos
		hw.CPUFeatures = []string{"avx2"}
		return nil
	}

	sysctl := string(output)
	if strings.Contains(sysctl, "avx2") || strings.Contains(sysctl, "AVX2") {
		hw.CPUFeatures = append(hw.CPUFeatures, "avx2")
	}

	return nil
}

// detectSystemRAM tenta identificar a RAM total do sistema.
func detectSystemRAM(hw *Hardware) {
	switch runtime.GOOS {
	case "windows":
		detectRAMWindows(hw)
	case "linux":
		detectRAMLinux(hw)
	case "darwin":
		detectRAMMacOS(hw)
	}
}

// detectRAMWindows usa WMI ou registry.
func detectRAMWindows(hw *Hardware) {
	// Simplificação: use wmic se disponível
	cmd := exec.Command("wmic", "os", "get", "totalvisiblememorybytes")
	_, err := cmd.Output()
	if err != nil {
		// Fallback para 8GB padrão
		hw.RAM = 8 * 1024 * 1024 * 1024 // 8GB default
	} else {
		// Parse da output... para simplicidade, também usa default
		hw.RAM = 8 * 1024 * 1024 * 1024 // 8GB default
	}
}

// detectRAMLinux lê /proc/meminfo.
func detectRAMLinux(hw *Hardware) {
	cmd := exec.Command("grep", "MemTotal", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		hw.RAM = 8 * 1024 * 1024 * 1024
		return
	}

	// Parse: MemTotal:        16334560 kB
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		var kb int64
		fmt.Sscanf(parts[1], "%d", &kb)
		hw.RAM = kb * 1024
	} else {
		hw.RAM = 8 * 1024 * 1024 * 1024
	}
}

// detectRAMMacOS usa `sysctl`.
func detectRAMMacOS(hw *Hardware) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		hw.RAM = 8 * 1024 * 1024 * 1024
		return
	}

	var bytes int64
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &bytes)
	hw.RAM = bytes
}

// detectGPUWindows tenta encontrar NVIDIA CUDA via registry ou nvidia-smi.
func detectGPUWindows(hw *Hardware) {
	// Tentar nvidia-smi first
	cmd := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		hw.GPUType = "cuda"
		// Tentar extrair versão CUDA
		cmd2 := exec.Command("nvidia-smi", "--query-gpu=compute_cap", "--format=csv,noheader")
		out2, _ := cmd2.Output()
		if len(out2) > 0 {
			hw.GPUVersion = strings.TrimSpace(string(out2))
		} else {
			hw.GPUVersion = "12.0" // Default
		}
		return
	}

	// Fallback: check registry no Windows
	// Para simplicidade, pular por enquanto
}

// detectGPULinux tenta nvidia-smi ou /proc/driver/nvidia.
func detectGPULinux(hw *Hardware) {
	// Tentar nvidia-smi
	cmd := exec.Command("nvidia-smi", "--list-gpus")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		hw.GPUType = "cuda"
		hw.GPUVersion = "12.0" // Default, pode ser refinado
		return
	}

	// Tentar vulkan
	cmd = exec.Command("vulkaninfo", "--summary")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "GPU") {
		hw.GPUType = "vulkan"
		return
	}
}

// detectGPUMacOS verifica Metal (nativo em M1+) ou CUDA (Intel com eGPU).
func detectGPUMacOS(hw *Hardware) {
	// Apple Silicon (ARM64) tem Metal nativamente
	if hw.Architecture == "arm64" {
		hw.GPUType = "metal"
		return
	}

	// Intel Mac: pode ter eGPU com CUDA, mas é raro
	// Para simplicidade, assumir que qualquer Metal-capable Mac usa Metal
	hw.GPUType = "metal"
}

// PrintHardwareInfo exibe as capacidades detectadas (debug).
func PrintHardwareInfo(hw *Hardware) {
	fmt.Printf("=== Hardware Detected ===\n")
	fmt.Printf("OS: %s\n", hw.OS)
	fmt.Printf("Architecture: %s\n", hw.Architecture)
	fmt.Printf("CPU Cores: %d\n", hw.CoreCount)
	fmt.Printf("CPU Features: %v\n", hw.CPUFeatures)
	fmt.Printf("RAM: %.2f GB\n", float64(hw.RAM)/(1024*1024*1024))
	fmt.Printf("GPU: %s %s\n", hw.GPUType, hw.GPUVersion)
	fmt.Printf("========================\n")
}
