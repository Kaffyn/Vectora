package desktop

import (
	"context"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Kaffyn/Vectora/desktop/ui"
)

// App gerencia a aplicação desktop
type App struct {
	fyneApp        fyne.App
	preferences    *Preferences
	ipcClient      *IPCClient
	indexClient    *IndexClient
	downloadMgr    *DownloadManager
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex

	// UI Components
	daemonStatusLabel string
	indexStatusLabel  string

	// State
	daemonConnected bool
	indexConnected  bool
}

// NewApp cria uma nova instância da aplicação desktop
func NewApp(fyneApp fyne.App, prefs *Preferences) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		fyneApp:             fyneApp,
		preferences:         prefs,
		ctx:                 ctx,
		cancel:              cancel,
		daemonConnected:     false,
		indexConnected:      false,
		daemonStatusLabel:   "○ Daemon",
		indexStatusLabel:    "○ Index Service",
	}

	// Inicializar IPC client (daemon local)
	ipcClient, ipcErr := NewIPCClient()
	if ipcErr != nil {
		log.Printf("[WARN] Não foi possível conectar ao daemon: %v\n", ipcErr)
	} else {
		app.ipcClient = ipcClient
		app.daemonConnected = true
		app.daemonStatusLabel = "● Daemon"
	}

	// Inicializar gRPC client (Index Service remoto)
	indexAddr := prefs.IndexServiceAddr
	if indexAddr == "" {
		indexAddr = "localhost:3000"
	}
	indexClient, grpcErr := NewIndexClient(indexAddr)
	if grpcErr != nil {
		log.Printf("[WARN] Não foi possível conectar ao Index Service: %v\n", grpcErr)
	} else {
		app.indexClient = indexClient
		app.indexConnected = true
		app.indexStatusLabel = "● Index Service"
	}

	// Inicializar Download Manager
	if app.indexClient != nil {
		app.downloadMgr = NewDownloadManager(app.indexClient, prefs)
	}

	// Iniciar monitor de conexão
	go app.monitorConnections()

	return app, nil
}

// BuildUI constrói a interface principal
func (a *App) BuildUI() fyne.CanvasObject {
	// Criar tabs principais com painéis da UI
	chatPanel := ui.NewChatPanel(a.ipcClient)
	editorPanel := ui.NewEditorPanel(a.ipcClient)
	indexPanel := ui.NewIndexPanel(a.ipcClient)
	settingsPanel := ui.NewSettingsPanel(a.ipcClient)

	tabs := container.NewAppTabs(
		container.NewTabItem("💬 Chat", chatPanel.GetContainer()),
		container.NewTabItem("💻 Code", editorPanel.GetContainer()),
		container.NewTabItem("📚 Index", indexPanel.GetContainer()),
		container.NewTabItem("⚙️ Settings", settingsPanel.GetContainer()),
	)

	// Status bar
	statusContainer := container.New(
		layout.NewHBoxLayout(),
		widget.NewLabel("Status: "),
		widget.NewLabel(a.daemonStatusLabel),
		widget.NewLabel(" | "),
		widget.NewLabel(a.indexStatusLabel),
		layout.NewSpacer(),
		widget.NewLabel("v1.0.0"),
	)

	// Container principal
	mainContainer := container.New(
		layout.NewBorderLayout(nil, statusContainer, nil, nil),
		tabs,
		statusContainer,
	)

	return mainContainer
}

// monitorConnections monitora a conexão com daemon e index service
func (a *App) monitorConnections() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			// Verificar daemon
			daemonOK := false
			if a.ipcClient != nil {
				err := a.ipcClient.EnsureConnected()
				daemonOK = (err == nil)
			}

			// Verificar Index Service
			indexOK := false
			if a.indexClient != nil {
				indexOK = a.indexClient.IsConnected()
			}

			a.setConnectionStatus(daemonOK, indexOK)
		}
	}
}

// setConnectionStatus atualiza o status de conexão
func (a *App) setConnectionStatus(daemon, index bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	changed := (a.daemonConnected != daemon) || (a.indexConnected != index)
	a.daemonConnected = daemon
	a.indexConnected = index

	if changed {
		if daemon {
			a.daemonStatusLabel = "● Daemon"
		} else {
			a.daemonStatusLabel = "○ Daemon"
		}
		if index {
			a.indexStatusLabel = "● Index Service"
		} else {
			a.indexStatusLabel = "○ Index Service"
		}
	}
}

// IsConnected retorna se está conectado a ambos os serviços
func (a *App) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.daemonConnected && a.indexConnected
}

// Close fecha recursos da aplicação
func (a *App) Close() error {
	a.cancel()

	if a.ipcClient != nil {
		a.ipcClient.Close()
	}

	if a.indexClient != nil {
		a.indexClient.Close()
	}

	if a.downloadMgr != nil {
		a.downloadMgr.Close()
	}

	return nil
}
