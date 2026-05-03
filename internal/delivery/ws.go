package delivery

import (
	"context"
	"discord/internal/core/domain"
	"encoding/json"
	"log"
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

func (sv *OLD_SERVER_HANDLERS) HandleConnections(w http.ResponseWriter, r *http.Request) {
	uidStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Println("Ошибка: не указан или неверен user_id")
		return
	}
	user_id := int64(userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// 2. Регистрируем нового клиента
	sv.clientsMu.Lock()
	sv.clients[user_id] = conn
	sv.clientsMu.Unlock()
	ctx := r.Context()

	log.Printf("Пользователь %d подключился!\n", userID)
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
		log.Println("Ошибка парсинга:", err)
		return
	}
	if input.Sender_id != actualSenderID {
		log.Printf("Внимание! Юзер %d пытается выдать себя за %d", actualSenderID, input.Sender_id)
		return
	}
	m := domain.Messages{
		Message:  input.Message,
		SenderId: input.Sender_id,
		ChatId:   input.Chat_id,
		Sent_at:  time.Now(),
	}
	if err := sv.service.SaveMessage(ctx, m); err != nil {
		log.Println("Ошибка БД:", err)
	}

	membersIDs, err := sv.service.GetChatMembers(ctx, m.ChatId)
	if err != nil {
		log.Println("Ошибка получения участников:", err)
		return
	}
	sv.clientsMu.Lock()
	for _, id := range membersIDs {
		if clientConn, ok := sv.clients[id]; ok {
			err := clientConn.WriteJSON(m)
			// err := clientConn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Ошибка отправки юзеру %d: %v", id, err)
				clientConn.Close()
				delete(sv.clients, id)
			}
		}
	}
	sv.clientsMu.Unlock()
}
