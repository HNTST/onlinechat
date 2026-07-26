package domain

import "time"

type Messages struct {
	Id        int64     `json:"id"`
	ChatId    int64     `json:"chat_id"`
	SenderId  int64     `json:"sender_id"`
	Message   string    `json:"message"`
	Sent_at   time.Time `json:"sent_at"`
	IsEdited  bool      `json:"is_edited"`
	IsDeleted bool      `json:"is_deleted"`
	IsRead    bool      `json:"is_read"`
}
