package desktop

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Kaffyn/Vectora/desktop/ui"
)

// App represents the main desktop application
type App struct {
	fyneApp fyne.App
	ipc     *IPCClient
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex

	// UI Components
	chatPanel     *ui.ChatPanel
	editorPanel   *ui.EditorPanel
	indexPanel    *ui.IndexPanel
	settingsPanel *ui.SettingsPanel
	statusBar     fyne.CanvasObject

	// State
	connected    bool
	activeModel  string
	activeIndices []string
}

// NewApp creates and initializes a new desktop application
func NewApp(fyneApp fyne.App) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	a := &App{
		fyneApp:       fyneApp,
		ctx:           ctx,
		cancel:        cancel,
		connected:     false,
		activeModel:   "",
		activeIndices: []string{},
	}

	// Initialize IPC client
	ipcClient, err := NewIPCClient()
	if err != nil {
		cancel()
		return nil, err
	}
	a.ipc = ipcClient

	// Initialize UI components
	a.chatPanel = ui.NewChatPanel(a.ipc)
	a.editorPanel = ui.NewEditorPanel(a.ipc)
	a.indexPanel = ui.NewIndexPanel(a.ipc)
	a.settingsPanel = ui.NewSettingsPanel(a.ipc)
	a.statusBar = a.buildStatusBar()

	// Start connection manager
	go a.manageConnection()

	return a, nil
}

// BuildUI constructs the main UI layout
func (a *App) BuildUI() fyne.CanvasObject {
	// Create tabs for different sections
	tabs := container.NewAppTabs()

	tabs.Append(container.NewTabItem("Chat", a.chatPanel.GetContainer()))
	tabs.Append(container.NewTabItem("Code Editor", a.editorPanel.GetContainer()))
	tabs.Append(container.NewTabItem("Index Manager", a.indexPanel.GetContainer()))
	tabs.Append(container.NewTabItem("Settings", a.settingsPanel.GetContainer()))

	// Create main layout: tabs + status bar
	mainContent := container.NewBorder(
		nil,                // top
		a.statusBar,        // bottom
		nil,                // left
		nil,                // right
		tabs,               // center
	)

	return mainContent
}

// buildStatusBar creates the status bar
func (a *App) buildStatusBar() fyne.CanvasObject {
	statusLabel := widget.NewLabel("Status: Initializing...")
	modelLabel := widget.NewLabel("Model: " + a.activeModel)

	return container.NewHBox(
		statusLabel,
		modelLabel,
	)
}

// manageConnection manages the IPC connection lifecycle
func (a *App) manageConnection() {
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			// Check connection status
			err := a.ipc.EnsureConnected()
			if err != nil {
				a.setConnected(false)
				// Retry connection after a delay
				select {
				case <-a.ctx.Done():
					return
				default:
					// Retry will happen on next iteration
				}
			} else {
				a.setConnected(true)
			}
		}
	}
}

// setConnected updates the connection status
func (a *App) setConnected(connected bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = connected
}

// IsConnected returns the current connection status
func (a *App) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// Close cleans up resources
func (a *App) Close() error {
	a.cancel()
	if a.ipc != nil {
		return a.ipc.Close()
	}
	return nil
}
