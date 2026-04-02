package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Para o servidor Vectora e o sidecar llama.cpp",
	Long: `Este comando é destinado a parar os componentes do Vectora que foram iniciados.
Atualmente, o comando 'start' roda em foreground e deve ser encerrado manualmente (Ctrl+C).
A funcionalidade de parar processos em background está em desenvolvimento.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Funcionalidade 'stop' em desenvolvimento.")
		fmt.Println("Se o servidor Vectora foi iniciado com 'start', por favor, encerre o processo manualmente (Ctrl+C).")
	},
}
