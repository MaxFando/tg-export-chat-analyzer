package parser

import (
	"time"
)

// Event represents a single message from Telegram export.
type Event struct {
	ID       int64     `json:"id"`
	FromID   string    `json:"from_id"` // Unique ID from JSON, used instead of username
	Text     string    `json:"text"`
	Date     time.Time `json:"date"`
	ChatID   int64     `json:"chat_id"` // ID of the chat
	Entities []Entity  `json:"entities,omitempty"`
}

// Entity represents text entities like mentions.
type Entity struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// MergedEvents holds a sorted list of unique events.
type MergedEvents struct {
	Events []Event `json:"events"`
}

// TelegramExport represents the root structure of Telegram JSON export
type TelegramExport struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	ID       int64        `json:"id"`
	Messages []RawMessage `json:"messages"`
}

// RawMessage is the raw structure from Telegram JSON export.
type RawMessage struct {
	ID       int64       `json:"id"`
	Type     string      `json:"type"`
	Date     string      `json:"date"`
	FromID   string      `json:"from_id"`
	From     string      `json:"from"`
	Text     interface{} `json:"text"`
	Entities []Entity    `json:"text_entities,omitempty"`
}
