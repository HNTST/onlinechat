package delivery

import (
	"context"
	"discord/internal/core/domain"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleConnections upgrades to WebSocket
//
//	@Summary		WebSocket-соединение
//	@Description	Обновление HTTP до WebSocket для обмена сообщениями в реальном времени
//	@Tags			v1 / WebSocket
//	@Param			user_id	query	int	true	"ID пользователя"
//	@Success		101		{object}	object{}
//	@Router			/api/v1/ws [get]
func (sv *OLD_SERVER_HANDLERS) HandleConnections(w http.ResponseWriter, r *http.Request) {
	uidStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(uidStr)
	if err != nil {
		sv.logger.Error("Ошибка: не указан или неверен user_id")
		return
	}
	user_id := int64(userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		sv.logger.Error("ошибка апгрейда до WebSocket", "error", err)
		return
	}
	defer conn.Close()

	sv.clientsMu.Lock()
	sv.clients[user_id] = conn
	sv.clientsMu.Unlock()
	ctx := r.Context()

	sv.logger.Info("Пользователь подключился", "user_id", userID)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			sv.clientsMu.Lock()
			delete(sv.clients, user_id)
			sv.clientsMu.Unlock()
			break
		}
		sv.ProcessMessage(ctx, msg, user_id)
	}
}

func (sv *OLD_SERVER_HANDLERS) ProcessMessage(ctx context.Context, data []byte, actualSenderID int64) {
	var input struct {
		Message   string `json:"message"`
		Chat_id   int64  `json:"chat_id"`
		Sender_id int64  `json:"sender_id"`
	}

	if err := json.Unmarshal(data, &input); err != nil {
		sv.logger.Error("Ошибка парсинга", "error", err)
		return
	}
	if input.Sender_id != actualSenderID {
		sv.logger.Warn("Попытка выдать себя за другого пользователя", "actual_sender_id", actualSenderID, "fake_sender_id", input.Sender_id)
		return
	}
	m := domain.Messages{
		Message:  input.Message,
		SenderId: input.Sender_id,
		ChatId:   input.Chat_id,
		Sent_at:  time.Now(),
	}
	if err := sv.service.SaveMessage(ctx, m); err != nil {
		sv.logger.Error("Ошибка БД", "error", err)
	}

	membersIDs, err := sv.service.GetChatMembers(ctx, m.ChatId)
	if err != nil {
		sv.logger.Error("Ошибка получения участников", "error", err)
		return
	}
	sv.clientsMu.Lock()
	for _, id := range membersIDs {
		if clientConn, ok := sv.clients[id]; ok {
			err := clientConn.WriteJSON(m)
			// err := clientConn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				sv.logger.Error("Ошибка отправки юзеру", "user_id", id, "error", err)
				clientConn.Close()
				delete(sv.clients, id)
			}
		}
	}
	sv.clientsMu.Unlock()
}
