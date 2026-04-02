package domain

import (
	"errors"
)

type Chunk struct {
        ID         string
        DocumentID string
        Content    string
        Index      int
        Metadata   map[string]string
}
func NewChunk(doc *Document, content string, index int) (*Chunk, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	return &Chunk{
		DocumentID: doc.ID,
		Content:    content,
		Index:      index,
	}, nil
}
