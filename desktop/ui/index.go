package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// IndexPanel represents the index manager interface
type IndexPanel struct {
	container *fyne.Container

	// Components
	indexList    *widget.List
	indexData    []IndexInfo
	statusLabel  *widget.Label
	addButton    *widget.Button
	removeButton *widget.Button
	refreshButton *widget.Button

	// State
	selectedIndex int
	ipcClient     interface{}
}

// IndexInfo represents index information
type IndexInfo struct {
	ID            string
	Name          string
	DocumentCount int
	Size          int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewIndexPanel creates a new index panel
func NewIndexPanel(ipcClient interface{}) *IndexPanel {
	ip := &IndexPanel{
		indexData:     []IndexInfo{},
		selectedIndex: -1,
		ipcClient:     ipcClient,
	}

	// Create index list
	ip.indexList = widget.NewList(
		func() int {
			return len(ip.indexData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(ip.indexData) {
				idx := ip.indexData[id]
				label.SetText(fmt.Sprintf("%s (%d docs, %.2f MB)",
					idx.Name, idx.DocumentCount, float64(idx.Size)/1024/1024))
			}
		},
	)
	ip.indexList.OnSelected = func(id widget.ListItemID) {
		ip.selectedIndex = id
	}

	// Create buttons
	ip.addButton = widget.NewButton("Add Index", func() {
		ip.handleAddIndex()
	})

	ip.removeButton = widget.NewButton("Remove Selected", func() {
		ip.handleRemoveIndex()
	})

	ip.refreshButton = widget.NewButton("Refresh", func() {
		ip.handleRefreshList()
	})

	// Create status label
	ip.statusLabel = widget.NewLabel("Ready")

	// Build toolbar
	toolbar := container.NewHBox(
		ip.addButton,
		ip.removeButton,
		ip.refreshButton,
		widget.NewSeparator(),
		ip.statusLabel,
	)

	// Build main layout
	ip.container = container.NewBorder(
		toolbar,      // top
		nil,          // bottom
		nil,          // left
		nil,          // right
		ip.indexList, // center
	)

	return ip
}

// GetContainer returns the Fyne container object
func (ip *IndexPanel) GetContainer() *fyne.Container {
	return ip.container
}

// handleAddIndex handles adding a new index
func (ip *IndexPanel) handleAddIndex() {
	ip.statusLabel.SetText("Add Index dialog would open here")
	// Placeholder - would open dialog in full implementation
}

// handleRemoveIndex handles removing the selected index
func (ip *IndexPanel) handleRemoveIndex() {
	if ip.selectedIndex < 0 {
		ip.statusLabel.SetText("No index selected")
		return
	}

	if ip.selectedIndex >= len(ip.indexData) {
		return
	}

	indexName := ip.indexData[ip.selectedIndex].Name
	ip.indexData = append(ip.indexData[:ip.selectedIndex], ip.indexData[ip.selectedIndex+1:]...)
	ip.indexList.Refresh()
	ip.statusLabel.SetText("Removed index: " + indexName)
	ip.selectedIndex = -1
}

// handleRefreshList handles refreshing the index list
func (ip *IndexPanel) handleRefreshList() {
	ip.statusLabel.SetText("Refreshing index list...")
	// In full implementation, would query daemon via IPC
	ip.statusLabel.SetText("Refresh complete")
}

// AddIndex adds an index to the list
func (ip *IndexPanel) AddIndex(info IndexInfo) {
	ip.indexData = append(ip.indexData, info)
	ip.indexList.Refresh()
}

// ClearIndices clears all indices
func (ip *IndexPanel) ClearIndices() {
	ip.indexData = []IndexInfo{}
	ip.indexList.Refresh()
}

// GetIndices returns the current index list
func (ip *IndexPanel) GetIndices() []IndexInfo {
	return ip.indexData
}
