package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// EditorPanel represents the code editor interface
type EditorPanel struct {
	container *fyne.Container

	// Components
	tabs          *container.AppTabs
	fileTree      *widget.List
	fileTreeData  []string
	editorContent *widget.Entry
	statusLabel   *widget.Label
	saveButton    *widget.Button

	// State
	currentFile  string
	openedFiles  map[string]string // filename -> content
	ipcClient    interface{}
}

// NewEditorPanel creates a new editor panel
func NewEditorPanel(ipcClient interface{}) *EditorPanel {
	ep := &EditorPanel{
		openedFiles: make(map[string]string),
		fileTreeData: []string{},
		ipcClient:   ipcClient,
	}

	// Create file tree
	ep.fileTree = widget.NewList(
		func() int {
			return len(ep.fileTreeData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(ep.fileTreeData[id])
		},
	)

	// Create editor content
	ep.editorContent = widget.NewEntry()
	ep.editorContent.Wrapping = false
	ep.editorContent.MultiLine = true

	// Create buttons
	saveButton := widget.NewButton("Save", func() {
		ep.handleSaveFile()
	})
	ep.saveButton = saveButton

	openButton := widget.NewButton("Open", func() {
		ep.handleOpenFile()
	})

	newButton := widget.NewButton("New", func() {
		ep.handleNewFile()
	})

	// Create status label
	ep.statusLabel = widget.NewLabel("No file open")

	// Build toolbar
	toolbar := container.NewHBox(
		newButton,
		openButton,
		saveButton,
		widget.NewSeparator(),
		ep.statusLabel,
	)

	// Build left panel (file tree)
	leftPanel := container.NewBorder(
		nil,          // top
		nil,          // bottom
		nil,          // left
		nil,          // right
		ep.fileTree,  // center
	)

	// Build right panel (editor)
	rightPanel := container.NewBorder(
		toolbar,           // top
		nil,               // bottom
		nil,               // left
		nil,               // right
		ep.editorContent,  // center
	)

	// Create main layout with split
	ep.container = container.NewHBox(
		container.NewBorder(
			nil, nil, nil, nil,
			leftPanel,
		),
		widget.NewSeparator(),
		container.NewBorder(
			nil, nil, nil, nil,
			rightPanel,
		),
	)

	return ep
}

// GetContainer returns the Fyne container object
func (ep *EditorPanel) GetContainer() *fyne.Container {
	return ep.container
}

// handleNewFile handles creating a new file
func (ep *EditorPanel) handleNewFile() {
	ep.currentFile = "untitled.txt"
	ep.editorContent.SetText("")
	ep.statusLabel.SetText("New file: " + ep.currentFile)
}

// handleOpenFile handles opening a file
func (ep *EditorPanel) handleOpenFile() {
	// Placeholder - would open file dialog in full implementation
	homeDir, _ := os.UserHomeDir()
	path := filepath.Join(homeDir, "example.txt")

	content, err := os.ReadFile(path)
	if err != nil {
		ep.statusLabel.SetText("Error opening file: " + err.Error())
		return
	}

	ep.currentFile = path
	ep.editorContent.SetText(string(content))
	ep.statusLabel.SetText("Opened: " + filepath.Base(path))
}

// handleSaveFile handles saving the current file
func (ep *EditorPanel) handleSaveFile() {
	if ep.currentFile == "" {
		ep.statusLabel.SetText("No file to save")
		return
	}

	content := ep.editorContent.Text
	err := os.WriteFile(ep.currentFile, []byte(content), 0644)
	if err != nil {
		ep.statusLabel.SetText("Error saving file: " + err.Error())
		return
	}

	ep.openedFiles[ep.currentFile] = content
	ep.statusLabel.SetText("Saved: " + filepath.Base(ep.currentFile))
}

// SetContent sets the editor content
func (ep *EditorPanel) SetContent(content string) {
	ep.editorContent.SetText(content)
}

// GetContent returns the editor content
func (ep *EditorPanel) GetContent() string {
	return ep.editorContent.Text
}
