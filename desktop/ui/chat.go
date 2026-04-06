package ui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ChatPanel represents the chat interface
type ChatPanel struct {
	container *fyne.Container

	// Components
	messageList  *widget.List
	messageData  []ChatMessage
	inputField   *widget.Entry
	sendButton   *widget.Button
	statusLabel  *widget.Label

	// State
	isLoading bool
	ipcClient interface{} // Will be IPCClient from parent
}

// ChatMessage represents a single chat message
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
	Model   string
}

// NewChatPanel creates a new chat panel
func NewChatPanel(ipcClient interface{}) *ChatPanel {
	cp := &ChatPanel{
		messageData: []ChatMessage{},
		ipcClient:   ipcClient,
	}

	// Create message list
	cp.messageList = widget.NewList(
		func() int {
			return len(cp.messageData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			msg := cp.messageData[id]
			label.SetText(fmt.Sprintf("[%s] %s", msg.Role, msg.Content))
		},
	)

	// Create input field
	cp.inputField = widget.NewEntry()
	cp.inputField.SetPlaceHolder("Type your message here...")
	cp.inputField.OnSubmitted = func(s string) {
		if cp.sendButton != nil {
			cp.sendButton.OnTapped()
		}
	}

	// Create send button
	cp.sendButton = widget.NewButton("Send", func() {
		cp.handleSendMessage()
	})

	// Create status label
	cp.statusLabel = widget.NewLabel("Ready")

	// Build layout
	inputBox := container.NewBorder(
		nil,           // top
		nil,           // bottom
		nil,           // left
		cp.sendButton, // right
		cp.inputField, // center
	)

	cp.container = container.NewBorder(
		nil,              // top
		inputBox,         // bottom
		nil,              // left
		nil,              // right
		cp.messageList,   // center
	)

	return cp
}

// GetContainer returns the Fyne container object
func (cp *ChatPanel) GetContainer() *fyne.Container {
	return cp.container
}

// handleSendMessage handles sending a chat message
func (cp *ChatPanel) handleSendMessage() {
	message := cp.inputField.Text
	if message == "" {
		return
	}

	// Clear input
	cp.inputField.SetText("")

	// Add user message
	cp.messageData = append(cp.messageData, ChatMessage{
		Role:    "user",
		Content: message,
	})
	cp.messageList.Refresh()

	// Update status
	cp.statusLabel.SetText("Sending...")
	cp.sendButton.Disable()
	cp.isLoading = true

	// Send message in background
	go cp.sendToDaemon(message)
}

// sendToDaemon sends message to daemon via IPC
func (cp *ChatPanel) sendToDaemon(message string) {
	defer func() {
		cp.sendButton.Enable()
		cp.isLoading = false
	}()

	// Placeholder implementation
	// In full implementation, this would use cp.ipcClient.Send()
	cp.messageData = append(cp.messageData, ChatMessage{
		Role:    "assistant",
		Content: "This is a placeholder response. Daemon integration pending.",
		Model:   "unknown",
	})
	cp.messageList.Refresh()

	cp.statusLabel.SetText("Ready")
}

// AddMessage adds a message to the chat
func (cp *ChatPanel) AddMessage(role, content, model string) {
	cp.messageData = append(cp.messageData, ChatMessage{
		Role:    role,
		Content: content,
		Model:   model,
	})
	cp.messageList.Refresh()
}

// ClearChat clears the message history
func (cp *ChatPanel) ClearChat() {
	cp.messageData = []ChatMessage{}
	cp.messageList.Refresh()
}

// GetMessages returns the message history
func (cp *ChatPanel) GetMessages() []ChatMessage {
	return cp.messageData
}
