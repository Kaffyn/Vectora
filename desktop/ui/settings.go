package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SettingsPanel represents the settings interface
type SettingsPanel struct {
	container *fyne.Container

	// Components - Models
	modelList      *widget.List
	modelData      []string
	activeModel    string
	modelLabel     *widget.Label

	// Components - UI/UX
	darkModeCheck *widget.Check
	fontSizeSelect *widget.Select

	// Components - System Info
	systemInfo     *widget.RichText

	// Buttons
	applyButton    *widget.Button
	resetButton    *widget.Button

	// State
	ipcClient      interface{}
}

// NewSettingsPanel creates a new settings panel
func NewSettingsPanel(ipcClient interface{}) *SettingsPanel {
	sp := &SettingsPanel{
		modelData:     []string{"gpt-4", "gpt-3.5-turbo", "claude-3"},
		activeModel:   "gpt-4",
		ipcClient:     ipcClient,
	}

	// Create model list
	sp.modelList = widget.NewList(
		func() int {
			return len(sp.modelData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(sp.modelData) {
				modelName := sp.modelData[id]
				active := ""
				if modelName == sp.activeModel {
					active = " (active)"
				}
				label.SetText(modelName + active)
			}
		},
	)

	// Create model label
	sp.modelLabel = widget.NewLabel("Active Model: " + sp.activeModel)

	// Create model section
	modelSection := container.NewVBox(
		widget.NewLabel("Available Models:"),
		sp.modelList,
		sp.modelLabel,
	)

	// Create UI/UX section
	sp.darkModeCheck = widget.NewCheck("Dark Mode", func(b bool) {
		// Handle dark mode toggle
	})

	sp.fontSizeSelect = widget.NewSelect(
		[]string{"Small", "Medium", "Large"},
		func(s string) {
			// Handle font size change
		},
	)
	sp.fontSizeSelect.SetSelected("Medium")

	uixSection := container.NewVBox(
		widget.NewLabel("Appearance:"),
		sp.darkModeCheck,
		widget.NewLabel("Font Size:"),
		sp.fontSizeSelect,
	)

	// Create system info
	sp.systemInfo = widget.NewRichTextFromMarkdown(
		"**System Information:**\n\n" +
		"Status: Connected\n\n" +
		"Uptime: 12h 34m\n\n" +
		"Models Path: ~/.vectora/models\n\n" +
		"Version: 1.0.0",
	)

	systemSection := container.NewVBox(
		widget.NewLabel("System Status:"),
		sp.systemInfo,
	)

	// Create buttons
	sp.applyButton = widget.NewButton("Apply", func() {
		sp.handleApply()
	})

	sp.resetButton = widget.NewButton("Reset", func() {
		sp.handleReset()
	})

	buttonBar := container.NewHBox(
		sp.applyButton,
		sp.resetButton,
	)

	// Create tabs for different settings sections
	tabs := container.NewAppTabs()
	tabs.Append(container.NewTabItem().
		Text("Models").
		Content(modelSection))
	tabs.Append(container.NewTabItem().
		Text("Appearance").
		Content(uixSection))
	tabs.Append(container.NewTabItem().
		Text("System").
		Content(systemSection))

	// Build main layout
	sp.container = container.NewBorder(
		nil,        // top
		buttonBar,  // bottom
		nil,        // left
		nil,        // right
		tabs,       // center
	)

	return sp
}

// GetContainer returns the Fyne container object
func (sp *SettingsPanel) GetContainer() *fyne.Container {
	return sp.container
}

// handleApply handles applying settings changes
func (sp *SettingsPanel) handleApply() {
	// In full implementation, would apply settings via daemon IPC
	fmt.Println("Settings applied")
}

// handleReset handles resetting settings to defaults
func (sp *SettingsPanel) handleReset() {
	sp.darkModeCheck.SetChecked(false)
	sp.fontSizeSelect.SetSelected("Medium")
	fmt.Println("Settings reset")
}

// SetActiveModel sets the active model and updates the display
func (sp *SettingsPanel) SetActiveModel(model string) {
	sp.activeModel = model
	sp.modelLabel.SetText("Active Model: " + model)
	sp.modelList.Refresh()
}

// GetActiveModel returns the currently selected active model
func (sp *SettingsPanel) GetActiveModel() string {
	return sp.activeModel
}

// SetModels sets the list of available models
func (sp *SettingsPanel) SetModels(models []string) {
	sp.modelData = models
	sp.modelList.Refresh()
}

// UpdateSystemInfo updates the system information display
func (sp *SettingsPanel) UpdateSystemInfo(status, uptime, modelsPath, version string) {
	markdown := fmt.Sprintf(
		"**System Information:**\n\n"+
		"Status: %s\n\n"+
		"Uptime: %s\n\n"+
		"Models Path: %s\n\n"+
		"Version: %s",
		status, uptime, modelsPath, version,
	)
	sp.systemInfo.ParseMarkdown(markdown)
}
