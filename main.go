package main

import (
	"discord/database"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"

	"github.com/gorilla/websocket"
)

// Upgrader нужен, чтобы превратить обычный HTTP запрос в WebSocket соединение
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключения со всех источников (для тестов в Postman)
	},
}

// Хранилище для всех активных соединений
// Используем Mutex, так как несколько горутин будут одновременно читать/писать в карту
var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 1. "Апгрейд" соединения
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 2. Регистрируем нового клиента
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	fmt.Println("Новый клиент подключился!")
	broadcast([]byte("Новый пользователь вошел в чат"))
	for {
		// 3. Читаем сообщение от клиента
		// ReadMessage блокирует выполнение горутины до поступления данных
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}

		// 4. Рассылаем полученное сообщение всем остальным (Broadcast)
		broadcast(msg)
	}
}

func broadcast(msg []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		response := fmt.Sprintf("Ты прислан мне [%s] и в нем количество байт = [%d]", string(msg), len(msg))
		err := client.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Предупреждение: .env не найден, используем системные переменные")
	}
	sv := database.Service{}
	_ = sv.InitDB()

	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Сервер запущен на :8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
