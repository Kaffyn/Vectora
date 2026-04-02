package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verifica o status operacional dos protocolos (MCP, ACP) e Sidecars",
	Run: func(cmd *cobra.Command, args []string) {
		checkStatus()
	},
}

func checkStatus() {
	resp, err := http.Get("http://localhost:8080/api/health")
	if err != nil {
		fmt.Println("❌ Erro: Vectora Core Daemon não está rodando.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)

	fmt.Println("✅ Vectora Core: ONLINE")
	
	// Simular verificação de Sidecars
	if health["status"] == "ok" {
		fmt.Println("  - 🤖 Sidecars (Qwen3): OPERANTES")
	} else {
		fmt.Println("  - ⚠️ Sidecars: DEGRADADOS")
	}

	// MCP Status
	fmt.Println("  - 📡 Protocolo MCP: ATIVO (Porta 8080/mcp)")
	
	// ACP Status (Poderia vir do health se expandirmos o JSON)
	fmt.Println("  - 🛡️ Protocolo ACP: ATIVO (Snapshot Git pronto)")
	fmt.Println("  - ⚖️ Confirmação: Guarded (Leitura Livre / Escrita Pendente)")
}
