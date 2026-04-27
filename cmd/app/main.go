package main

import (
	"discord/internal/config"
	"discord/internal/delivery"
	"discord/internal/repository"
	"discord/internal/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	repo := repository.NewDatabse(cfg.Config_DB)
	if err := repo.InitDB(); err != nil {
		log.Println("Ошибка подлючения к базе данных")
		return
	}
	defer repo.Close()

	service := service.NewService(repo)

	app := delivery.NewServer(service)

	mux := http.NewServeMux() // Рекомендуется использовать свой Mux вместо глобального
	mux.HandleFunc("/ws", app.HandleConnections())
	mux.HandleFunc("POST /chat", app.SaveChat())
	mux.HandleFunc("POST /user", app.SaveUser())

	mux.HandleFunc("GET /user_id", app.GetUserID())
	mux.HandleFunc("GET /user", app.GetUser())
	mux.HandleFunc("GET /chats", app.GetUserChats())
	mux.HandleFunc("GET /messages", app.GetUserMessages())

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
