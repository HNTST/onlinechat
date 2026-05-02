package main

import (
	"context"
	core_config "discord/internal/core/config"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	core_http_server "discord/internal/core/transport/server"
	"discord/internal/delivery"
	"discord/internal/repository/minio"
	"discord/internal/repository/postgres"
	"discord/internal/service"
	"os/signal"
	"syscall"
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

	sv := service.NewService(repo, mn)
	oldTransport := delivery.NewServer(logger, sv)
	oldRoutes := oldTransport.Routes()
	apiVersionRouter := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion1)
	// <--->
	// pgPool, err := core_postgres_pool.NewConnectionPool(ctx, cfg.Config_DB)
	// if err != nil {
	// 	logger.Error("PgPool run error", "error", err)
	// }

	// usersRepo := users_postgres_repository.NewUsersRepository(pgPool)
	// usersService := users_service.NewUsersService(usersRepo)
	// usersTransport := users_transport_http.NewUsersHTTPHandler(usersService)
	// usersRoutes := usersTransport.Routes()
	// apiVersionRouter2 := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion2)

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
	// httpServer.RegisterAPIRoutes(apiVersionRouter2)
	// apiVersionRouter2.RegisterRoutes(usersRoutes...)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", "error", err)
	}

}
