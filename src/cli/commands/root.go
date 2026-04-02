package commands

import (

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vectora-cli",
	Short: "Vectora CLI v2.0 - Deep Technical Memory",
	Long: `Vectora CLI fornece uma interface de linha de comando para interagir com o Vectora,
um assistente de IA focado em memória técnica profunda para Desenvolvimento de Jogos.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Aqui você pode definir flags globais e lógica de inicialização.
	// Por exemplo: rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "arquivo de configuração (default é $HOME/.vectora-cli.yaml)")
	// rootCmd.PersistentFlags().Bool("verbose", false, "saída detalhada")
}
