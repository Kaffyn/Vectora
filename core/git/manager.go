package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Manager handles Git operations for snapshots.
type Manager struct {
	WorkDir   string
	Available bool
}

func NewManager(workDir string) *Manager {
	m := &Manager{WorkDir: workDir}
	m.checkGit()
	return m
}

func (m *Manager) checkGit() {
	if _, err := exec.LookPath("git"); err != nil {
		m.Available = false
		return
	}

	cmd := exec.Command("git", "-C", m.WorkDir, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		m.Available = false
		return
	}
	m.Available = true
}

// Snapshot creates a git snapshot of a specific file.
func (m *Manager) Snapshot(filePath string) error {
	if !m.Available {
		return fmt.Errorf("git not available")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Check if file is tracked or modified
	statusCmd := exec.Command("git", "-C", m.WorkDir, "status", "--porcelain", absPath)
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		return nil // Not in git repo, silently skip
	}
	if len(output) == 0 {
		return nil // File not modified, no snapshot needed
	}

	// Add specific file only
	addCmd := exec.Command("git", "-C", m.WorkDir, "add", absPath)
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	commitCmd := exec.Command("git", "-C", m.WorkDir, "commit", "-m",
		fmt.Sprintf("chore(vectora): snapshot pre-edit [%s]", time.Now().Format(time.RFC3339)))
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

// Undo reverts the last Vectora snapshot.
func (m *Manager) Undo() error {
	if !m.Available {
		return fmt.Errorf("git not available")
	}

	// Find last Vectora commit
	logCmd := exec.Command("git", "-C", m.WorkDir, "log", "-1", "--oneline", "--grep=chore(vectora)")
	output, err := logCmd.CombinedOutput()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("no Vectora snapshot found")
	}

	// Reset to before that commit
	resetCmd := exec.Command("git", "-C", m.WorkDir, "reset", "--soft", "HEAD^")
	return resetCmd.Run()
}

// IsGitRepo checks if the work directory is a git repository.
func (m *Manager) IsGitRepo() bool {
	return m.Available
}

// Init initializes a git repository if one doesn't exist.
func (m *Manager) Init() error {
	if m.Available {
		return nil
	}

	cmd := exec.Command("git", "-C", m.WorkDir, "init")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Create initial commit
	cmd = exec.Command("git", "-C", m.WorkDir, "add", ".")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", m.WorkDir, "commit", "-m", "Initial commit")
	_ = cmd.Run()

	m.Available = true
	return nil
}

// CreateRepo initializes a .gitignore if it doesn't exist.
func (m *Manager) CreateGitignore(entries ...string) error {
	gitignorePath := filepath.Join(m.WorkDir, ".gitignore")

	// Check if exists
	if _, err := os.Stat(gitignorePath); err == nil {
		return nil // Already exists
	}

	content := "# Vectora managed files\n"
	for _, e := range entries {
		content += e + "\n"
	}

	return os.WriteFile(gitignorePath, []byte(content), 0644)
}
