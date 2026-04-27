package service

import (
	dm "discord/internal/domain"
)

type Service struct {
	db dm.Repository
}

func NewService(db dm.Repository) *Service {
	return &Service{
		db: db,
	}
}
func (sv *Service) SaveUser(username, password string) (int64, error) {
	return sv.db.SaveUser(username, password)
}
func (sv *Service) SaveChat(chat_name string, user_id, target_user_id int64) (int64, error) {
	return sv.db.SaveChatWithAuthor(chat_name, user_id, target_user_id)
}

func (sv *Service) GetUser(user_id int64) (dm.User, error) {
	return sv.db.GetUser(user_id)
}
func (sv *Service) GetUserChats(user_id int64) ([]dm.Chat, error) {
	return sv.db.GetUserChats(user_id)
}
func (sv *Service) GetUserMessages(chat_id int64) ([]dm.Messages, error) {
	return sv.db.GetChatMessages(chat_id)
}
func (sv *Service) GetUserID(username string) (int64, error) {
	return sv.db.GetUserID(username)
}
func (sv *Service) GetChatMembers(chat_id int64) ([]int64, error) {
	return sv.db.GetChatMembers(chat_id)
}
func (sv *Service) SaveMessage(message dm.Messages) error {
	return sv.db.SaveMessage(message)
}
