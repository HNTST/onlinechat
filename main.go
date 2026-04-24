package main

import (
	"database/sql"
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

	dbRepo := db.Service{Stmt: make(map[string]*sql.Stmt)}
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

	mux.HandleFunc("GET /user_id", serv.GetUserID())
	mux.HandleFunc("GET /user", serv.GetUser())
	mux.HandleFunc("GET /chats", serv.GetUserChats())
	mux.HandleFunc("GET /messages", serv.GetUserMessages())

	handlerWithCORS := enableCORS(mux)

	fmt.Println("Сервер запущен на :8080/ws")
	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithCORS,
	}

	log.Fatal(server.ListenAndServe())
}
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого источника (*)
		// Для продакшена лучше указать конкретный адрес фронтенда друга
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Разрешаем методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		// Разрешаем заголовки (Content-Type важен для JSON)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если это preflight-запрос (OPTIONS), просто отвечаем 200 OK и выходим
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
