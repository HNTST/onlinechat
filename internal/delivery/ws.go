package delivery

import (
	"discord/internal/domain"
	md "discord/internal/domain"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (sv *Server) HandleConnections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uidStr := r.URL.Query().Get("user_id")
		userID, err := strconv.Atoi(uidStr)
		if err != nil {
			sv.sendError(w, domain.ErrParamMissing.Error(), http.StatusBadRequest, nil)
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

		sv.logger.Info("Пользователь подлючился", "user_id", userID)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				sv.clientsMu.Lock()
				delete(sv.clients, user_id)
				sv.clientsMu.Unlock()
				break
			}
			sv.ProcessMessage(msg, user_id)
		}
	}
}

func (sv *Server) ProcessMessage(data []byte, actualSenderID int64) {
	var m md.Messages
	if err := json.Unmarshal(data, &m); err != nil {
		log.Println("Ошибка парсинга:", err)
		return
	}
	if m.SenderId != actualSenderID {
		log.Printf("Внимание! Юзер %d пытается выдать себя за %d", actualSenderID, m.SenderId)
		return
	}
	if err := sv.service.SaveMessage(m); err != nil {
		log.Println("Ошибка БД:", err)
	}

	membersIDs, err := sv.service.GetChatMembers(m.ChatId)
	if err != nil {
		log.Println("Ошибка получения участников:", err)
		return
	}
	sv.clientsMu.Lock()
	for _, id := range membersIDs {
		if clientConn, ok := sv.clients[id]; ok {
			err := clientConn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Ошибка отправки юзеру %d: %v", id, err)
				clientConn.Close()
				delete(sv.clients, id)
			}
		}
	}
	sv.clientsMu.Unlock()
}
