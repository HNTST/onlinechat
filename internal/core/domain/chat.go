package domain

import "time"

type Chat struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	IsGroup bool   `json:"is_group,omitempty"`
}

type MessageInfo struct {
	Id       int64     `json:"id"`
	Text     string    `json:"text"`
	SenderId int64     `json:"sender_id"`
	SentAt   time.Time `json:"sent_at"`
	IsRead   bool      `json:"is_read"`
}

type ChatInfo struct {
	Id          int64        `json:"id"`
	Name        string       `json:"name"`
	IsGroup     bool         `json:"is_group,omitempty"`
	LastMessage *MessageInfo `json:"last_message,omitempty"`
	UnreadCount int          `json:"unread_count"`
}
