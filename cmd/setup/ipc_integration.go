package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kaffyn/Vectora/internal/ipc"
)

// IPCModelInstaller gerencia instalação de modelos via IPC
type IPCModelInstaller struct {
	client *ipc.Client
}

// NewIPCModelInstaller cria nova instância
func NewIPCModelInstaller() (*IPCModelInstaller, error) {
	client, err := ipc.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC client: %w", err)
	}

	// Tentar conectar com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("failed to connect to daemon: %w", err)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("connection timeout: daemon not running")
	}

	return &IPCModelInstaller{client: client}, nil
}

// InstallModel instala um modelo via IPC
func (im *IPCModelInstaller) InstallModel(modelID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	payload := map[string]string{"model_id": modelID}

	var response map[string]interface{}
	err := im.client.Send(ctx, "model.install", payload, &response)
	if err != nil {
		return fmt.Errorf("install request failed: %w", err)
	}

	return nil
}

// SetActiveModel define modelo ativo via IPC
func (im *IPCModelInstaller) SetActiveModel(modelID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := map[string]string{"model_id": modelID}

	var response map[string]interface{}
	err := im.client.Send(ctx, "model.set-active", payload, &response)
	if err != nil {
		return fmt.Errorf("set-active request failed: %w", err)
	}

	return nil
}

// GetRecommendedModel retorna modelo recomendado
func (im *IPCModelInstaller) GetRecommendedModel() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var model map[string]interface{}
	err := im.client.Send(ctx, "model.recommend", json.RawMessage(`{}`), &model)
	if err != nil {
		return "", fmt.Errorf("recommend request failed: %w", err)
	}

	if id, ok := model["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("invalid recommend response")
}

// ListModels lista modelos disponíveis
func (im *IPCModelInstaller) ListModels() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var models []map[string]interface{}
	err := im.client.Send(ctx, "model.list", json.RawMessage(`{}`), &models)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}

	return models, nil
}

// Close fecha a conexão IPC
func (im *IPCModelInstaller) Close() {
	if im.client != nil {
		im.client.Close()
	}
}
