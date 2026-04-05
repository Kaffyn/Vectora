package main

import (
	"context"
	"encoding/json"
)

// App represents the Wails application bindings
type App struct {
	ipcClient interface{} // Will be initialized in main
	ctx       context.Context
}

// CallIPC calls the daemon via IPC protocol
func (a *App) CallIPC(method string, payload json.RawMessage) (json.RawMessage, error) {
	// This will be implemented to call the actual IPC client
	// For now, return a placeholder response
	return json.RawMessage(`{"status":"ok"}`), nil
}

// OnEvent subscribes to daemon events
func (a *App) OnEvent(eventName string, callback func(json.RawMessage)) {
	// This will be implemented to subscribe to actual events
}
