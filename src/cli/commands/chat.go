package commands

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Inicia uma sessão de chat interativa com o Vectora",
	Long:  "Este comando inicia um terminal interativo para conversar com o Vectora. Ele se conecta ao servidor HTTP do Vectora para enviar suas perguntas e receber respostas.",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Ocorreu um erro ao iniciar o chat: %v\n", err)
			os.Exit(1)
		}
	},
}

// Model representa o estado do nosso programa Bubble Tea.
type model struct {
	textInput string
	history   []string
}

// initialModel retorna um modelo inicial.
func initialModel() model {
	return model{
		textInput: "",
		history:   []string{"Bem-vindo ao chat com o Vectora! Digite sua pergunta e pressione Enter."},
	}
}

// Init é chamada uma vez no início do programa.
func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen // Entra no modo Alt Screen do terminal.
}

// Update processa eventos e atualiza o modelo.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Quando Enter é pressionado, adiciona o texto ao histórico e limpa o input.
			if m.textInput != "" {
				m.history = append(m.history, "> Você: "+m.textInput)
				// Por enquanto, uma resposta dummy.
				m.history = append(m.history, "Vectora: Processando sua requisição...")
				m.textInput = ""
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit // Sai do programa.
		case tea.KeyBackspace:
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
		case tea.KeyRunes: // Para caracteres regulares
			m.textInput += msg.String()
		}
	}
	return m, nil
}

// View renderiza a interface do usuário.
func (m model) View() string {
	// A View() será mais complexa com quebras de linha, então vou forçar a concatenação manual.
	s := "Histórico do Chat:\n"
	for _, line := range m.history {
		s += line + "\n"
	}
	s += "----------------------------------------\n"
	s += "Sua pergunta: " + m.textInput + "\n"
	s += "(Pressione Ctrl+C ou Esc para sair)"
	return s
}
