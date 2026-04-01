package desktop

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// UploadSession representa uma sessão de upload
type UploadSession struct {
	SessionID      string
	IndexID        string
	WorkspaceID    string
	UploadURL      string
	Status         string // pending, uploading, processing, completed, failed
	Progress       int    // 0-100
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ErrorMessage   string
	DocumentCount  int
}

// DownloadManager gerencia uploads de documentos via navegador
type DownloadManager struct {
	indexClient   *IndexClient
	preferences   *Preferences
	sessions      map[string]*UploadSession
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	pollTicker    *time.Ticker
}

// NewDownloadManager cria um novo gerenciador de downloads
func NewDownloadManager(indexClient *IndexClient, prefs *Preferences) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())

	mgr := &DownloadManager{
		indexClient: indexClient,
		preferences: prefs,
		sessions:    make(map[string]*UploadSession),
		ctx:         ctx,
		cancel:      cancel,
		pollTicker:  time.NewTicker(5 * time.Second),
	}

	// Iniciar monitor de sessões
	go mgr.monitorSessions()

	return mgr
}

// CreateUploadSession cria uma nova sessão de upload e abre no navegador
func (m *DownloadManager) CreateUploadSession(ctx context.Context, indexID, workspaceID string, documentCount int) (*UploadSession, error) {
	// Criar sessão no Index Service
	uploadURL, err := m.indexClient.CreateUploadSession(ctx, indexID, workspaceID, documentCount)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão de upload: %w", err)
	}

	// Extrair session ID da URL
	parsedURL, _ := url.Parse(uploadURL)
	sessionID := parsedURL.Path[len("/upload/"):]

	session := &UploadSession{
		SessionID:     sessionID,
		IndexID:       indexID,
		WorkspaceID:   workspaceID,
		UploadURL:     uploadURL,
		Status:        "pending",
		Progress:      0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DocumentCount: documentCount,
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	// Abrir URL no navegador padrão
	m.openInBrowser(uploadURL)

	return session, nil
}

// openInBrowser abre uma URL no navegador padrão
func (m *DownloadManager) openInBrowser(urlStr string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("start", urlStr)
	case "darwin":
		cmd = exec.Command("open", urlStr)
	case "linux":
		cmd = exec.Command("xdg-open", urlStr)
	default:
		return fmt.Errorf("sistema operacional não suportado: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// GetSession obtém uma sessão por ID
func (m *DownloadManager) GetSession(sessionID string) *UploadSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// ListSessions lista todas as sessões
func (m *DownloadManager) ListSessions() []*UploadSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*UploadSession
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// monitorSessions monitora o status das sessões de upload
func (m *DownloadManager) monitorSessions() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.pollTicker.C:
			m.mu.RLock()
			sessions := make(map[string]*UploadSession)
			for k, v := range m.sessions {
				sessions[k] = v
			}
			m.mu.RUnlock()

			for sessionID, session := range sessions {
				// Pular sessões completadas ou com erro
				if session.Status == "completed" || session.Status == "failed" {
					continue
				}

				// Verificar status no Index Service
				status, err := m.indexClient.GetUploadStatus(m.ctx, sessionID)
				if err != nil {
					session.Status = "failed"
					session.ErrorMessage = err.Error()
					continue
				}

				// Atualizar status
				if newStatus, ok := status["status"].(string); ok {
					session.Status = newStatus
				}
				if progress, ok := status["progress"].(float64); ok {
					session.Progress = int(progress)
				}
				session.UpdatedAt = time.Now()
			}
		}
	}
}

// RetryUpload redefine uma sessão para retry
func (m *DownloadManager) RetryUpload(sessionID string) error {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("sessão não encontrada: %s", sessionID)
	}

	if session.Status != "failed" {
		return fmt.Errorf("apenas sessões com erro podem ser retomadas")
	}

	// Abrir URL novamente no navegador
	m.openInBrowser(session.UploadURL)

	// Resetar status
	m.mu.Lock()
	session.Status = "pending"
	session.Progress = 0
	session.ErrorMessage = ""
	session.UpdatedAt = time.Now()
	m.mu.Unlock()

	return nil
}

// CancelSession cancela uma sessão de upload
func (m *DownloadManager) CancelSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("sessão não encontrada: %s", sessionID)
	}

	if session.Status == "completed" {
		return fmt.Errorf("não é possível cancelar upload completado")
	}

	delete(m.sessions, sessionID)
	return nil
}

// ClearCompletedSessions remove sessões completadas da lista
func (m *DownloadManager) ClearCompletedSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for sessionID, session := range m.sessions {
		if session.Status == "completed" {
			delete(m.sessions, sessionID)
		}
	}
}

// Close fecha o gerenciador de downloads
func (m *DownloadManager) Close() error {
	m.cancel()
	if m.pollTicker != nil {
		m.pollTicker.Stop()
	}
	return nil
}
