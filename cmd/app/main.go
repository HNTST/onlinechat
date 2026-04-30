package main

import (
	"context"
	"discord/internal/config"
	"discord/internal/delivery"
	"discord/internal/repository/minio"
	"discord/internal/repository/postgres"
	"discord/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	cfg := config.Load()

	logger.Info("Подлючение к базе данных posgres")
	repo := postgres.NewDatabse(cfg.Config_DB)
	if err := repo.InitDB(); err != nil {
		logger.Error("Ошибка подлючения к базе данных", "error", err.Error())
		return
	}
	logger.Info("Успешное подлючение к базе  postges")

	logger.Info("Подлючение к базе данных minio")
	mn := minio.NewMinio(cfg.Config_Minio)
	if err := mn.InitMinio(); err != nil {
		logger.Error("Ошибка подлючения к базе данных", "error", err.Error())
		return
	}
	logger.Info("Успешное подлючение к базе minio")
	service := service.NewService(repo, mn)

	app := delivery.NewServer(logger, service)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", app.HandleConnections())
	mux.HandleFunc("POST /chat", app.SaveChat())
	mux.HandleFunc("POST /user", app.SaveUser())
	mux.HandleFunc("POST /user/avatar", app.UpdateUserAvatar())

	mux.HandleFunc("GET /user_id", app.GetUserID())
	mux.HandleFunc("GET /user", app.GetUser())
	mux.HandleFunc("GET /chats", app.GetUserChats())
	mux.HandleFunc("GET /messages", app.GetUserMessages())

	handlerWithMVChain := app.MWChain(mux, app.CORSMV, app.LoggingMV)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithMVChain,
	}

	go func() {
		logger.Info("Сервер запущен")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Ошибка запуска сервера", "error", err.Error())
			return
		}
	}()
	<-signalChan
	logger.Info("Получен сигнал завершения, начинаю graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("Завершение работы сервера")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при завершении сервера", "error", err.Error())
		return
	}
	logger.Info("Закрытие соединения с базой данных")
	if err := repo.Close(); err != nil {
		logger.Error("Ошибка при завершнии работы с базой данных", "error", err.Error())
		return
	}

	logger.Info("Сервер завершил работу")

}
