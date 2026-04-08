package telemetry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultMaxSize = 10 * 1024 * 1024 // 10 MB
	BackupSuffix   = ".old"
)

// RotatingWriter implementa io.Writer com rotação simples de arquivos.
type RotatingWriter struct {
	filepath string
	maxSize  int64
	current  *os.File
	mu       sync.Mutex
}

func NewRotatingWriter(path string, maxSize int64) (*RotatingWriter, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}

	rw := &RotatingWriter{
		filepath: path,
		maxSize:  maxSize,
	}

	// Garante que o diretório existe
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, err
	}

	// Abre ou cria o arquivo inicial
	if err := rw.openFile(); err != nil {
		return nil, err
	}

	return rw, nil
}

func (rw *RotatingWriter) openFile() error {
	f, err := os.OpenFile(rw.filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	rw.current = f
	return nil
}

// rotate executa a lógica de renomear o atual para .old e criar um novo.
func (rw *RotatingWriter) rotate() error {
	if rw.current != nil {
		rw.current.Close()
	}

	backupPath := rw.filepath + BackupSuffix

	// Remove o backup antigo se existir (mantém apenas 1 geração)
	os.Remove(backupPath)

	// Renomeia o atual para backup
	if err := os.Rename(rw.filepath, backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to rotate log: %w", err)
	}

	// Cria novo arquivo limpo
	return rw.openFile()
}

func (rw *RotatingWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Verifica tamanho atual
	info, err := rw.current.Stat()
	if err != nil {
		return 0, err
	}

	// Se escrevermos p, ultrapassaremos o limite?
	if info.Size()+int64(len(p)) > rw.maxSize {
		if err := rw.rotate(); err != nil {
			return 0, err
		}
	}

	return rw.current.Write(p)
}

func (rw *RotatingWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.current != nil {
		return rw.current.Close()
	}
	return nil
}
