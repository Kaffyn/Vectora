package domain

import (
	"time"
)

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewConversation(id string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        id,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Conversation) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	c.UpdatedAt = time.Now()
}
