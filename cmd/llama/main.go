package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	silent := flag.Bool("silent", false, "Executa em modo silencioso (sem interface)")
	path := flag.String("path", "", "Pasta de destino para o motor Llama-CPP")
	flag.Parse()

	// If path not provided, use current
	installDir := *path
	if installDir == "" {
		installDir = "./llama-engine"
	}

	if !*silent {
		fmt.Println("==============================================")
		fmt.Printf(" LLAMA-CPP ENGINE INSTALLER (%s)\n", runtime.GOOS)
		fmt.Println("==============================================")
	}

	// Create destination folder
	_ = os.MkdirAll(installDir, 0755)

	assets := getLlamaAssets()
	if len(assets) == 0 {
		fmt.Printf("[ERRO] Nenhum ativo encontrado para %s. Recompile o projeto.\n", runtime.GOOS)
		os.Exit(1)
	}

	for filename, data := range assets {
		target := filepath.Join(installDir, filename)
		if !*silent {
			fmt.Printf("[INSTALANDO] %-20s (%d bytes)... ", filename, len(data))
		}
		
		err := os.WriteFile(target, data, 0755)
		if err != nil {
			if !*silent { fmt.Printf("FALHA: %v\n", err) }
			os.Exit(1)
		}
		if !*silent { fmt.Println("OK") }
	}

	if !*silent {
		fmt.Println("\nInstalação do motor local concluída com sucesso em:", installDir)
	}
}
