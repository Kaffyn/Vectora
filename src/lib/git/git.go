package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitBridge fornece métodos seguros para o ACP interagir com o repositório do usuário.
type GitBridge struct {
	workDir string
}

func NewGitBridge(workDir string) *GitBridge {
	return &GitBridge{workDir: workDir}
}

// CreateSnapshot cria um commit temporário para permitir rollback fácil pelo ACP.
func (g *GitBridge) CreateSnapshot(message string) (string, error) {
	// 1. git add .
	if err := g.run("add", "."); err != nil {
		return "", fmt.Errorf("git add failed: %w", err)
	}

	// 2. git commit -m "[ACP] ..."
	commitMsg := fmt.Sprintf("[ACP-Snap] %s", message)
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = g.workDir
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		// Se não houver nada para commitar, não é um erro fatal
		if strings.Contains(string(out), "nothing to commit") {
			return "no-change", nil
		}
		return "", fmt.Errorf("git commit failed: %s (err: %w)", string(out), err)
	}

	// 3. Pegar o hash do commit
	hash, err := g.getOutput("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(hash), nil
}

// Rollback restaura o estado para um hash específico.
func (g *GitBridge) Rollback(hash string) error {
	return g.run("reset", "--hard", hash)
}

// GetDiff retorna a diferença atual das modificações não commitadas.
func (g *GitBridge) GetDiff() (string, error) {
	return g.getOutput("diff")
}

func (g *GitBridge) run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	return cmd.Run()
}

func (g *GitBridge) getOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
