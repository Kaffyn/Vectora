package domain

import (
	"errors"
	"time"
)

type Document struct {
	ID        string
	FilePath  string
	Content   string
	Type      string
	CreatedAt time.Time
}

func NewDocument(path, content, docType string) (*Document, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}

	return &Document{
		FilePath:  path,
		Content:   content,
		Type:      docType,
		CreatedAt: time.Now(),
	}, nil
}
