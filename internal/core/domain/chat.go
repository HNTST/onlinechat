package domain

type Chat struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// type ChatRepository interface {
// 	GetUserChats(user_id int64) ([]Chat, error)
// 	GetChatMembers(chat_id int64) ([]int64, error)
// 	SaveUserToChat(user_id, chat_id int64) error
// 	SaveChatWithAuthor(name string, user_id, target_user_id int64) (int64, error)
// }
// type ChatService interface {
// 	SaveChat(chat_name string, user_id, target_user_id int64) (int64, error)
// 	GetUserChats(user_id int64) ([]Chat, error)
// 	GetChatMembers(chat_id int64) ([]int64, error)
// }
