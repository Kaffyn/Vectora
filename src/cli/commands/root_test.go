package commands_test

import (
	"bytes"
	"testing"

	"github.com/Kaffyn/Vectora/src/cli/commands"
)

func TestCLI_300Percent(t *testing.T) {
	// 1. HAPPY PATH: Executar o comando sem argumentos (deve mostrar ajuda ou erro controlado)
	t.Run("HappyPath_HelpDisplay", func(t *testing.T) {
		root := commands.NewRootCmd()
		b := bytes.NewBufferString("")
		root.SetOut(b)
		root.SetArgs([]string{"--help"})

		if err := root.Execute(); err != nil {
			t.Errorf("Execução falhou: %v", err)
		}

		if !bytes.Contains(b.Bytes(), []byte("Vectora")) {
			t.Error("Ajuda não exibiu o nome da aplicação")
		}
	})

	// 2. NEGATIVE: Comando inexistente
	t.Run("Negative_UnknownCommand", func(t *testing.T) {
		root := commands.NewRootCmd()
		root.SetArgs([]string{"missing"})
		if err := root.Execute(); err == nil {
			t.Error("Esperava erro para comando inexistente, got nil")
		}
	})

	// ...
}
