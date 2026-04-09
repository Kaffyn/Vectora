// Package conversations manages per-workspace conversation persistence.
// Conversations are stored as JSONL files in ~/.Vectora/conversations/<session-id>.jsonl
// The session ID is deterministically derived from the workspace path (sha256).
package conversations

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry is a single message stored in a conversation file.
type Entry struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"ts"`
}

// Meta contains summary information about a stored conversation.
type Meta struct {
	ID        string
	Path      string
	Lines     int
	UpdatedAt time.Time
}

// SessionID derives a stable, deterministic conversation ID from an absolute workspace path.
// Format: ws_<first-8-hex-chars-of-sha256(path)>
func SessionID(absPath string) string {
	h := sha256.Sum256([]byte(filepath.ToSlash(strings.ToLower(absPath))))
	return fmt.Sprintf("ws_%x", h[:4])
}

// Store handles all persistence operations for a conversations directory.
type Store struct {
	dir string
}

// NewStore creates a Store backed by the given directory (created on demand).
func NewStore(baseDir string) *Store {
	return &Store{dir: filepath.Join(baseDir, "conversations")}
}

// filePath returns the JSONL path for a given session ID.
func (s *Store) filePath(sessionID string) string {
	return filepath.Join(s.dir, sessionID+".jsonl")
}

// Append writes a single message entry to the session file.
func (s *Store) Append(sessionID, role, content string) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}

	entry := Entry{Role: role, Content: content, Timestamp: time.Now()}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.filePath(sessionID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(string(line) + "\n")
	return err
}

// Load reads all entries from a session file. Returns nil if the file does not exist.
func (s *Store) Load(sessionID string) ([]Entry, error) {
	data, err := os.ReadFile(s.filePath(sessionID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []Entry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e Entry
		if json.Unmarshal([]byte(line), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

// Clear deletes the session file, effectively resetting the conversation.
func (s *Store) Clear(sessionID string) error {
	err := os.Remove(s.filePath(sessionID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// List returns metadata for all stored conversations in the directory.
func (s *Store) List() ([]Meta, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []Meta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, _ := e.Info()
		id := strings.TrimSuffix(e.Name(), ".jsonl")
		path := filepath.Join(s.dir, e.Name())

		data, _ := os.ReadFile(path)
		lines := strings.Count(string(data), "\n")

		meta := Meta{ID: id, Path: path, Lines: lines}
		if info != nil {
			meta.UpdatedAt = info.ModTime()
		}
		result = append(result, meta)
	}
	return result, nil
}
