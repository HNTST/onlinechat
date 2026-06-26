package main

// @title Discord API
// @version 1.0
// @description API мессенджер / чат-приложение (v1 legacy + v2 feature-based)
// @BasePath /
import (
	"context"
	_ "discord/docs"
	core_config "discord/internal/core/config"
	core_logger "discord/internal/core/logger"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	"discord/internal/delivery"
	chats_postgres_repository "discord/internal/features/chat/repository/postgres"
	chat_service "discord/internal/features/chat/service"
	chat_transport_http "discord/internal/features/chat/transport/http"
	chat_transport_ws "discord/internal/features/chat/transport/ws"
	files_repository "discord/internal/features/files/repository"
	files_postgres_repository "discord/internal/features/files/repository/postgres"
	files_service "discord/internal/features/files/service"
	files_transport_http "discord/internal/features/files/transport/http"
	users_postgres_repository "discord/internal/features/user/repository/postges"
	users_service "discord/internal/features/user/service"
	users_transport_http "discord/internal/features/user/transport/http"
	"discord/internal/repository/minio"
	"discord/internal/repository/postgres"
	"discord/internal/service"
	"net/http"
	"os/signal"
	"syscall"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger := core_logger.NewLogger()
	cfg := core_config.Load()

	// Старые обработчики переписанные под новый HTTP
	logger.Info("Подлючение к базе данных posgres")
	pgPool, err := core_postgres_pool.NewConnectionPool(ctx, cfg.Config_DB)
	if err != nil {
		logger.Error("PgPool run error", "error", err)
	}
	logger.Info("Успешное подлючение к базе  postges")

	logger.Info("Подлючение к базе данных minio")
	mn := minio.NewMinio(cfg.Config_Minio)
	if err := mn.InitMinio(); err != nil {
		logger.Error("Ошибка подлючения к базе данных", "error", err.Error())
		return
	}
	logger.Info("Успешное подлючение к базе minio")

	oldRepo := postgres.NewOldRepostiroty(pgPool)
	sv := service.NewService(oldRepo, mn)
	oldTransport := delivery.NewServer(logger, sv)
	oldRoutes := oldTransport.Routes()
	apiVersionRouter := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion1)
	// <--->

	// --- V2 User module ---
	usersRepo := users_postgres_repository.NewUsersRepository(pgPool)
	usersService := users_service.NewUsersService(usersRepo)
	usersTransport := users_transport_http.NewUsersHTTPHandler(usersService)
	usersRoutes := usersTransport.Routes()

	// --- V2 Chat module ---
	chatRepo := chats_postgres_repository.NewChatRepository(pgPool)
	chatService := chat_service.NewChatService(chatRepo)
	chatTransport := chat_transport_http.NewChatHttpHandler(chatService)
	chatRoutes := chatTransport.Routes()
	chatWSHandler := chat_transport_ws.NewChatWSHandler(chatService)

	// --- V2 File module ---
	filePostgresRepo := files_postgres_repository.NewFileRepository(pgPool)
	fileStorage, err := files_repository.NewMinioStorage(cfg.Config_Minio)
	if err != nil {
		logger.Error("Minio storage init error", "error", err)
		return
	}
	fileService := files_service.NewFileService(filePostgresRepo, fileStorage)
	fileTransport := files_transport_http.NewFileHTTPHandler(fileService)
	fileRoutes := fileTransport.Routes()

	apiVersionRouter2 := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion2)

	httpServer := core_http_server.NewHTTPServer(
		cfg.Config_Server,
		logger,
		core_http_middleware.Cors(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)

	// тоже для старых обработчиков
	httpServer.RegisterAPIRoutes(apiVersionRouter)
	apiVersionRouter.RegisterRoutes(oldRoutes...)
	// <--->

	httpServer.RegisterAPIRoutes(apiVersionRouter2)
	apiVersionRouter2.RegisterRoutes(usersRoutes...)
	apiVersionRouter2.RegisterRoutes(chatRoutes...)
	apiVersionRouter2.RegisterRoutes(fileRoutes...)
	apiVersionRouter2.RegisterRoutes(
		core_http_server.NewRoute(http.MethodGet, "/ws", chatWSHandler.HandleConnections),
	)

	httpServer.Handle("/swagger/", httpSwagger.WrapHandler)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", "error", err)
	}

}
