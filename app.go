package main

import (
	"context"
	"fmt"
	"github.com/Kaffyn/vectora/internal/ipc"
)

// App struct
type App struct {
	ctx        context.Context
	ipcClient  *ipc.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initializes the IPC client to connect with the Daemon
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Printf("Falha ao criar cliente IPC: %v\n", err)
		return
	}
	
	// Attempt to connect (Silent Daemon boot should be managed by the caller or here)
	if err := client.Connect(); err != nil {
		fmt.Printf("Aviso: Daemon offline no boot do App. %v\n", err)
	}
	
	a.ipcClient = client
}

// CallIPC sends a call to the Daemon via IPC
func (a *App) CallIPC(method string, payload string) (string, error) {
	if a.ipcClient == nil {
		return "", fmt.Errorf("daemon_offline")
	}

	// Legacy route mapping coming from the frontend
	routeMap := map[string]string{
		"/api/chat":          "workspace.query",
		"/api/conversations": "chat.list",
		"/api/settings":      "provider.get",
		"/api/search":        "memory.search",
	}

	if target, ok := routeMap[method]; ok {
		method = target
	}

	var result []byte
	err := a.ipcClient.SendRaw(context.Background(), method, []byte(payload), &result)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
