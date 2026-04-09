package main

// This file is a thin CLI adapter for core/conversations.
// All business logic lives in core/conversations/store.go.

import (
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/core/conversations"
)

// convStore is the shared instance for the CLI process.
var convStore *conversations.Store

func init() {
	userProfile, _ := os.UserHomeDir()
	convStore = conversations.NewStore(filepath.Join(userProfile, ".Vectora"))
}

func workspaceConversationID(absPath string) string {
	return conversations.SessionID(absPath)
}

func appendConversationEntry(sessionID, role, content string) error {
	return convStore.Append(sessionID, role, content)
}

func loadConversationEntries(sessionID string) ([]conversations.Entry, error) {
	return convStore.Load(sessionID)
}

func clearConversation(sessionID string) error {
	return convStore.Clear(sessionID)
}
