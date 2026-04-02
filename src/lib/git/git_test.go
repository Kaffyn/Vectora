package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/src/lib/git"
)

// setupRepo cria um repositório git temporário para testes.
func setupRepo(t *testing.T) string {
	dir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatal(err)
	}

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Criar arquivo inicial
	f := filepath.Join(dir, "initial.txt")
	os.WriteFile(f, []byte("hello"), 0644)
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Initial commit")

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
}

// TestGitBridge_HappyPath (100%): Cenário de sucesso de Snapshot e Rollback.
func TestGitBridge_HappyPath(t *testing.T) {
	dir := setupRepo(t)
	defer os.RemoveAll(dir)

	g := git.NewGitBridge(dir)

	// Alterar arquivo
	f := filepath.Join(dir, "initial.txt")
	os.WriteFile(f, []byte("modified"), 0644)

	// Snapshot
	hash, err := g.CreateSnapshot("Modification test")
	if err != nil {
		t.Errorf("CreateSnapshot failed: %v", err)
	}

	if hash == "" || hash == "no-change" {
		t.Errorf("Expected valid hash, got %s", hash)
	}

	// Verificar se committou
	diff, _ := g.GetDiff()
	if diff != "" {
		t.Errorf("Expected clean diff after snapshot, got: %s", diff)
	}

	// Rollback? (Poderia testar o rollback para o inicial aqui se tivesse guardado o hash inicial)
}

// TestGitBridge_Negative (200%): Testar comportamento com falhas.
func TestGitBridge_Negative(t *testing.T) {
	// 1. Diretório não git
	dir, _ := os.MkdirTemp("", "no-git")
	defer os.RemoveAll(dir)

	g := git.NewGitBridge(dir)
	_, err := g.CreateSnapshot("Fail test")
	if err == nil {
		t.Error("Expected error when running GitBridge in a non-git directory")
	}

	// 2. Rollback hash inválido
	dirGit := setupRepo(t)
	defer os.RemoveAll(dirGit)
	gGit := git.NewGitBridge(dirGit)

	errRoll := gGit.Rollback("invalid_hash_xyz")
	if errRoll == nil {
		t.Error("Expected error when rolling back to invalid hash")
	}
}

// TestGitBridge_EdgeCase (300%): Cenários limite (Arquivo bloqueado, git lock, etc).
func TestGitBridge_EdgeCase(t *testing.T) {
	dir := setupRepo(t)
	defer os.RemoveAll(dir)

	g := git.NewGitBridge(dir)

	// 1. Snapshot sem alterações
	hash, err := g.CreateSnapshot("Nothing to change")
	if err != nil {
		t.Fatalf("Expected no error on empty snapshot, got: %v", err)
	}
	if hash != "no-change" {
		t.Errorf("Expected 'no-change' status, got: %s", hash)
	}

	// 2. Snapshot de arquivos muito grandes ou binários (Simulado apenas como path longo)
	longPath := filepath.Join(dir, "very/long/path/to/some/file.txt")
	os.MkdirAll(filepath.Dir(longPath), 0755)
	os.WriteFile(longPath, []byte("content"), 0644)

	_, errLong := g.CreateSnapshot("Long path test")
	if errLong != nil {
		t.Errorf("Failed to snapshot long path: %v", errLong)
	}
}
