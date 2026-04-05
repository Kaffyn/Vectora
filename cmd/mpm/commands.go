package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Kaffyn/Vectora/internal/models"
)

// handleList lista todos os modelos disponíveis
func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	family := fs.String("family", "", "Filter by family")

	fs.Parse(args)

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	catalog := mm.GetCatalog()
	var filtered []models.Model

	if *family != "" {
		for _, m := range catalog.Models {
			if strings.Contains(m.ID, *family) {
				filtered = append(filtered, m)
			}
		}
	} else {
		filtered = catalog.Models
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(filtered, "", "  ")
		fmt.Println(string(data))
	} else {
		for _, m := range filtered {
			status := ""
			if mm.IsInstalled(m.ID) {
				status = " [INSTALLED]"
			}
			fmt.Printf("  %s - %s (%dB)%s\n", m.ID, m.Name, m.SizeBytes, status)
		}
	}
}

// handleDetect detecta hardware do sistema
func handleDetect(args []string) {
	fs := flag.NewFlagSet("detect", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Parse(args)

	hw, err := models.DetectHardware()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(hw, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("OS:          %s\n", hw.OS)
		fmt.Printf("Architecture: %s\n", hw.Architecture)
		fmt.Printf("CPU Cores:   %d\n", hw.CoreCount)
		fmt.Printf("CPU Features: %v\n", hw.CPUFeatures)
		fmt.Printf("RAM:         %.2f GB\n", float64(hw.RAM)/(1024*1024*1024))
		fmt.Printf("GPU Type:    %s\n", hw.GPUType)
		fmt.Printf("GPU Version: %s\n", hw.GPUVersion)
	}
}

// handleRecommend recomenda um modelo baseado no hardware
func handleRecommend(args []string) {
	fs := flag.NewFlagSet("recommend", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	sizeOverride := fs.String("size", "", "Override size recommendation")

	fs.Parse(args)

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	hw, err := models.DetectHardware()
	if err != nil {
		fmt.Printf("Error detecting hardware: %v\n", err)
		os.Exit(1)
	}

	var recommended *models.Model

	if *sizeOverride != "" {
		recommended, err = models.FindModel(mm.GetCatalog(), fmt.Sprintf("qwen3-%s", *sizeOverride))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		recommended, err = mm.RecommendModel(hw)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(recommended, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Recommended Model: %s\n", recommended.Name)
		fmt.Printf("ID:                %s\n", recommended.ID)
		fmt.Printf("Size:              %.2f GB\n", float64(recommended.SizeBytes)/1024/1024/1024)
		fmt.Printf("Required RAM:      %.2f GB\n", recommended.RequiredRAMGB)
		fmt.Printf("Required VRAM:     %.2f GB\n", recommended.RequiredVRAMGB)
	}
}

// handleSearch busca modelos
func handleSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Parse(args)
	remaining := fs.Args()

	if len(remaining) == 0 {
		fmt.Println("Usage: mpm search <query>")
		os.Exit(1)
	}

	query := remaining[0]

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	results := models.SearchByName(mm.GetCatalog(), query)

	if *jsonOutput {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(results) == 0 {
			fmt.Println("No models found matching your query")
		} else {
			for _, m := range results {
				fmt.Printf("  %s - %s\n", m.ID, m.Name)
			}
		}
	}
}

// handleInstall instala um modelo
func handleInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	model := fs.String("model", "", "Model ID to install")

	fs.Parse(args)

	if *model == "" {
		fmt.Println("Usage: mpm install --model <model-id>")
		os.Exit(1)
	}

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Installing %s...\n", *model)

	// Esta é uma implementação stub - download real seria implementado depois
	err = mm.Install(nil, *model, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Model %s registered (download simulation)\n", *model)
}

// handleActive mostra o modelo ativo
func handleActive(args []string) {
	fs := flag.NewFlagSet("active", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Parse(args)

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	activeID := mm.GetActive()
	if activeID == "" {
		fmt.Println("No active model set")
		return
	}

	model, err := models.FindModel(mm.GetCatalog(), activeID)
	if err != nil {
		fmt.Printf("Active model not found in catalog: %s\n", activeID)
		return
	}

	if *jsonOutput {
		data, _ := json.MarshalIndent(model, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Active Model: %s (%s)\n", model.Name, model.ID)
	}
}

// handleSetActive define o modelo ativo
func handleSetActive(args []string) {
	fs := flag.NewFlagSet("set-active", flag.ExitOnError)
	model := fs.String("model", "", "Model ID to set as active")

	fs.Parse(args)

	if *model == "" {
		fmt.Println("Usage: mpm set-active --model <model-id>")
		os.Exit(1)
	}

	mm, err := models.NewModelManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = mm.SetActive(*model)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Model %s set as active\n", *model)
}
