package main

import (
	"context"
	"discord/internal/config"
	"discord/internal/delivery"
	"discord/internal/repository"
	"discord/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	cfg := config.Load()

	repo := repository.NewDatabse(cfg.Config_DB)
	if err := repo.InitDB(); err != nil {
		log.Println("Ошибка подлючения к базе данных")
		return
	}

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

	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithCORS,
	}

	go func() {
		fmt.Println("Сервер запущен на :8080/ws")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("[ERROR] Init Server: ", err)
			return
		}
	}()
	<-signalChan
	fmt.Print("\nПолучен сигнал завершения, начинаю graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Print("\nЗавершение работы сервера")
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении сервера: %v", err)
		return
	}
	fmt.Println("\nЗакрытие соединения с базой данных")
	if err := repo.Close(); err != nil {
		fmt.Printf("Ошибка при завершнии работы с базой данных: %v", err)
		return
	}

	fmt.Println("\nСервер завершил работу")

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
