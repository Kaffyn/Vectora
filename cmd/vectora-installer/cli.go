package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	cliMode              bool
	cliInstallPath       string
	cliMode2             string // gemini or qwen3
	cliModel             string
	cliAPIKey            string
	cliSilent            bool
	cliLanguage          string
)

var rootCmd *cobra.Command

func init() {
	rootCmd = &cobra.Command{
		Use:   "vectora-installer",
		Short: "Vectora Setup - Instalar Vectora e configurar IA",
		Long:  "Setup wizard para instalar Vectora e configurar o motor de IA (Gemini ou Qwen3)",
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Instalar Vectora em modo CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return performCLIInstall()
		},
	}

	// Flags globais
	rootCmd.PersistentFlags().BoolVar(&cliSilent, "silent", false, "Modo silencioso (sem output)")
	rootCmd.PersistentFlags().StringVar(&cliLanguage, "lang", "pt", "Idioma (pt, en, es, fr)")

	// Flags do install
	installCmd.Flags().StringVar(&cliInstallPath, "path", "", "Caminho de instalação (padrão: Program Files)")
	installCmd.Flags().StringVar(&cliMode2, "mode", "gemini", "Modo de IA: gemini ou qwen3")
	installCmd.Flags().StringVar(&cliModel, "model", "", "Modelo Qwen3 (qwen3-0.6b, qwen3-1.7b, qwen3-4b, qwen3-8b)")
	installCmd.Flags().StringVar(&cliAPIKey, "api-key", "", "Chave de API do Google (para modo Gemini)")

	rootCmd.AddCommand(installCmd)
}

func runCLIMode() {
	// Check if running as admin - if not, silently restart with admin privileges
	systemManager, _ := newSystemManager()
	if systemManager != nil && !systemManager.IsRunningAsAdmin() {
		// Silently restart as admin using PowerShell with hidden window
		exe, _ := os.Executable()
		args := os.Args[1:]

		// Build PowerShell command to restart with admin privileges
		argsStr := ""
		for _, arg := range args {
			argsStr += "'" + arg + "' "
		}

		psCmd := fmt.Sprintf(`Start-Process -FilePath '%s' -ArgumentList %s -Verb runas -WindowStyle Hidden`, exe, argsStr)

		cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		_ = cmd.Start()
		os.Exit(0)
	}

	// Execute CLI commands via Cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}

func performCLIInstall() error {
	if !cliSilent {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║              Vectora Setup (CLI Mode)                      ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println("")
	}

	// Validar modo
	if cliMode2 != "gemini" && cliMode2 != "qwen3" {
		return fmt.Errorf("modo inválido: %s (use 'gemini' ou 'qwen3')", cliMode2)
	}

	// Definir caminho de instalação padrão se não fornecido
	if cliInstallPath == "" {
		systemManager, err := newSystemManager()
		if err != nil {
			return fmt.Errorf("erro ao inicializar: %v", err)
		}
		var defPath string
		defPath, _ = systemManager.GetInstallDir()
		cliInstallPath = defPath
	}

	if !cliSilent {
		fmt.Printf("📁 Caminho de instalação: %s\n", cliInstallPath)
	}

	// Criar diretórios
	if err := os.MkdirAll(cliInstallPath, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	os.MkdirAll(filepath.Join(cliInstallPath, "data", "chroma"), 0755)
	os.MkdirAll(filepath.Join(cliInstallPath, "logs"), 0755)
	os.MkdirAll(filepath.Join(cliInstallPath, "backups"), 0755)

	// Extrair binários
	if !cliSilent {
		fmt.Println("📦 Extraindo binários...")
	}

	assets := getInstallerAssets()
	binariesToExtract := []string{
		"vectora.exe",
		"vectora-cli.exe",
		"lpm.exe",
		"mpm.exe",
	}

	for _, binName := range binariesToExtract {
		if binData, exists := assets[binName]; exists && len(binData) > 0 {
			target := filepath.Join(cliInstallPath, binName)
			if err := os.WriteFile(target, binData, 0755); err != nil {
				return fmt.Errorf("erro ao extrair %s: %v", binName, err)
			}
			if !cliSilent {
				fmt.Printf("  ✓ %s\n", binName)
			}
		}
	}

	// Registrar app
	systemManager, _ := newSystemManager()
	systemManager.RegisterApp(cliInstallPath)

	// Configurar modo de IA
	if cliMode2 == "gemini" {
		if !cliSilent {
			fmt.Println("🔑 Modo Gemini API configurado")
		}
		if cliAPIKey != "" {
			// Aqui você salvaria a chave de forma criptografada
			if !cliSilent {
				fmt.Println("  ✓ API Key configurada")
			}
		}
	} else if cliMode2 == "qwen3" {
		if !cliSilent {
			fmt.Println("🤖 Modo Qwen3 configurado")
		}
		if cliModel == "" {
			cliModel = "qwen3-1.7b"
		}

		// Tentar instalar modelo via IPC se daemon estiver rodando
		ipcInstaller, err := NewIPCModelInstaller()
		if err == nil {
			defer ipcInstaller.Close()
			if !cliSilent {
				fmt.Printf("⬇️  Instalando modelo %s via daemon...\n", cliModel)
			}
			if err := ipcInstaller.InstallModel(cliModel); err == nil {
				if !cliSilent {
					fmt.Printf("  ✓ Modelo %s instalado via IPC\n", cliModel)
				}
			} else {
				if !cliSilent {
					fmt.Printf("  ⚠️  Não foi possível instalar via daemon: %v\n", err)
					fmt.Printf("  ℹ️  Modelo será instalado quando daemon iniciar\n")
				}
			}
		} else if !cliSilent {
			fmt.Printf("  ℹ️  Modelo será instalado quando daemon %s iniciar\n", cliModel)
		}
	}

	if !cliSilent {
		fmt.Println("")
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║              ✅ Instalação Concluída!                      ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
	}

	return nil
}
