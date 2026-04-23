package main

import (
	db "discord/database"
	sv "discord/server"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Предупреждение: .env не найден, используем системные переменные")
	}

	dbRepo := db.Service{}
	err = dbRepo.InitDB()
	if err != nil {
		log.Println("Ошибка подлючения к базе данных")
		return
	}

	serv := sv.New(&dbRepo)
	mux := http.NewServeMux() // Рекомендуется использовать свой Mux вместо глобального
	mux.HandleFunc("/ws", serv.HandleConnections())
	mux.HandleFunc("POST /chat", serv.SaveChat())
	mux.HandleFunc("POST /user", serv.SaveUser())
	mux.HandleFunc("GET /user", serv.GetUser())
	mux.HandleFunc("GET /chats", serv.GetUserChats())
	fmt.Println("Сервер запущен на :8080/ws")
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
