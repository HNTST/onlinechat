package main

// @title Discord API
// @version 2.0
// @description API мессенджер / чат-приложение (v2 feature-based)
// @BasePath /
import (
	"context"
	_ "discord/docs"
	core_config "discord/internal/core/config"
	core_logger "discord/internal/core/logger"
	core_postgres_pool "discord/internal/core/repository/postgres/pool"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	auth_postgres_repository "discord/internal/features/auth/repository/postgres"
	auth_service "discord/internal/features/auth/service"
	auth_transport_http "discord/internal/features/auth/transport/http"
	chats_postgres_repository "discord/internal/features/chat/repository/postgres"
	chat_service "discord/internal/features/chat/service"
	chat_transport_http "discord/internal/features/chat/transport/http"
	chat_transport_ws "discord/internal/features/chat/transport/ws"
	files_repository "discord/internal/features/files/repository"
	files_postgres_repository "discord/internal/features/files/repository/postgres"
	files_service "discord/internal/features/files/service"
	files_transport_http "discord/internal/features/files/transport/http"
	users_postgres_repository "discord/internal/features/user/repository/postgres"
	users_service "discord/internal/features/user/service"
	users_transport_http "discord/internal/features/user/transport/http"
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

	logger.Info("Подлючение к базе данных postgres")
	pgPool, err := core_postgres_pool.NewConnectionPool(ctx, cfg.Config_DB)
	if err != nil {
		logger.Error("PgPool run error", "error", err)
	}
	logger.Info("Успешное подлючение к базе postgres")

	logger.Info("Запуск миграций")
	if err := core_postgres_pool.RunMigrations(pgPool); err != nil {
		logger.Error("Migration error", "error", err)
		return
	}
	logger.Info("Миграции успешно применены")

	// --- V2 User module ---
	usersRepo := users_postgres_repository.NewUsersRepository(pgPool)
	usersService := users_service.NewUsersService(usersRepo)
	usersTransport := users_transport_http.NewUsersHTTPHandler(usersService)
	usersRoutes := usersTransport.Routes()

	// --- V2 Auth module ---
	authRepo := auth_postgres_repository.NewAuthRepository(pgPool)
	authService := auth_service.NewAuthService(authRepo, usersService)
	authTransport := auth_transport_http.NewAuthHTTPHandler(authService)
	authRoutes := authTransport.Routes()

	// --- V2 Chat module ---
	chatRepo := chats_postgres_repository.NewChatRepository(pgPool)
	chatService := chat_service.NewChatService(chatRepo)
	chatTransport := chat_transport_http.NewChatHttpHandler(chatService)
	chatRoutes := chatTransport.Routes()
	chatWSHandler := chat_transport_ws.NewChatWSHandler(ctx, logger, chatService, usersService)

	// --- V2 File module ---
	filePostgresRepo := files_postgres_repository.NewFileRepository(pgPool)
	logger.Info("Подлючение к базе данных Minio")
	fileStorage, err := files_repository.NewMinioStorage(cfg.Config_Minio)
	if err != nil {
		logger.Error("Minio storage init error", "error", err)
		return
	}
	logger.Info("Успешное подключение к MinIO")
	fileService := files_service.NewFileService(filePostgresRepo, fileStorage, usersService)
	fileTransport := files_transport_http.NewFileHTTPHandler(fileService)
	fileRoutes := fileTransport.Routes()

	apiVersionRouter := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion2)

	httpServer := core_http_server.NewHTTPServer(
		cfg.Config_Server,
		logger,
		core_http_middleware.Cors(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
		core_http_middleware.Auth(authService.ValidateSession),
	)

	httpServer.RegisterAPIRoutes(apiVersionRouter)
	apiVersionRouter.RegisterRoutes(usersRoutes...)
	apiVersionRouter.RegisterRoutes(chatRoutes...)
	apiVersionRouter.RegisterRoutes(fileRoutes...)
	apiVersionRouter.RegisterRoutes(authRoutes...)
	apiVersionRouter.RegisterRoutes(
		core_http_server.NewRoute(http.MethodGet, "/ws", chatWSHandler.HandleConnections),
	)

	httpServer.Handle("/swagger/", httpSwagger.WrapHandler)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", "error", err)
	}
}
