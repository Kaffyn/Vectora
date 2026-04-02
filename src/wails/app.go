package wails

import (
	"context"

	"github.com/Kaffyn/Vectora/src/core/ai"
	"github.com/Kaffyn/Vectora/src/core/config"
)

// App struct
type App struct {
	ctx          context.Context
	modelManager *ai.ModelManager
	paths        config.VectoraPaths
}

// NewApp creates a new App application struct
func NewApp() *App {
	paths := config.GetDefaultPaths()
	return &App{
		paths:        paths,
		modelManager: ai.NewModelManager(paths),
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.paths.EnsureDirectories()
}

// GetModelsRegistry retorna os modelos disponíveis no backserver.
func (a *App) GetModelsRegistry() (*ai.RegistryResponse, error) {
	return a.modelManager.FetchRegistry("http://localhost:3000/v1/registry")
}

// GetLibraryRegistry retorna os data packs disponíveis no backserver.
func (a *App) GetLibraryRegistry() (*ai.LibraryResponse, error) {
	return a.modelManager.FetchLibrary("http://localhost:3000/v1/library")
}

// InstallModel aciona o download de um modelo GGUF.
func (a *App) InstallModel(info ai.ModelInfo) error {
	return a.modelManager.InstallModel(info, func(p float64) {
		// runtime.EventsEmit(a.ctx, "download_progress", info.ID, p)
	})
}

// InstallDataPack aciona o download de um pack de treinamento vetorial.
func (a *App) InstallDataPack(info ai.DataPackInfo) error {
	return a.modelManager.InstallDataPack(info)
}

// UninstallModel remove um modelo local.
func (a *App) UninstallModel(id string) error {
	return a.modelManager.RemoveModel(id)
}

// GetInstalledModels retorna a lista de IDs de modelos instalados.
func (a *App) GetInstalledModels() ([]string, error) {
	return a.modelManager.ListInstalledModels()
}

// GetConfigPaths retorna os caminhos atuais do sistema (para depuração/exibição).
func (a *App) GetConfigPaths() config.VectoraPaths {
	return a.paths
}
