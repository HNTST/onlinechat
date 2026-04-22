package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	sv "discord/database"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	fmt.Println("Новый клиент подлючился!")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}
		broadcast(msg)
	}
}
func broadcast(msg []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
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
	serv := &sv.Service{}
	err = serv.InitDB()
	if err != nil {
		log.Printf("Ошибка подлючения к базе данных: %v", err)
		return
	}

	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Сервер запущен на :8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
