package server

import (
	md "discord/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (sv *Server) HandleConnections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		// 2. Регистрируем нового клиента
		sv.clientsMu.Lock()
		sv.clients[conn] = true
		sv.clientsMu.Unlock()

		fmt.Println("Новый клиент подключился!")
		sv.Broadcast([]byte("Новый пользователь вошел в чат"))
		for {
			// 3. Читаем сообщение от клиента
			// ReadMessage блокирует выполнение горутины до поступления данных
			_, msg, err := conn.ReadMessage()
			if err != nil {
				sv.clientsMu.Lock()
				delete(sv.clients, conn)
				sv.clientsMu.Unlock()
				break
			}

			sv.Broadcast(msg)

			go func(data []byte) {
				var m md.Messages
				if err := json.Unmarshal(data, &m); err != nil {
					log.Println("Ошибка парсинга сообщения:", err)
					return
				}
				if err := sv.repo.SaveMessage(m); err != nil {
					log.Println("Ошибка сохранения сообщения:", err)
				}
			}(msg)
		}
	}
}

func (sv *Server) Broadcast(msg []byte) {
	sv.clientsMu.Lock()
	defer sv.clientsMu.Unlock()

	for client := range sv.clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			client.Close()
			delete(sv.clients, client)
		}
	}
}
