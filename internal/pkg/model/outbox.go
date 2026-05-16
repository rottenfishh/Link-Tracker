package model

import (
	"time"
)

type Outbox struct {
	ID          string    `json:"id"`
	Topic       string    `json:"topic"`
	MessageType string    `json:"message_type"`
	Payload     []byte    `json:"payload"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessedAt time.Time `json:"processed_at"`
}
