package core_http_server

import (
	"context"
	core_config "discord/internal/core/config"
	core_logger "discord/internal/core/logger"
	core_http_middleware "discord/internal/core/transport/middleware"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	mux *http.ServeMux
	cfg core_config.Config_Server
	log *core_logger.Logger

	middleware []core_http_middleware.Middleware
}

func NewHTTPServer(cfg core_config.Config_Server, log *core_logger.Logger, middleware ...core_http_middleware.Middleware) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		cfg:        cfg,
		log:        log,
		middleware: middleware,
	}
}

func (h *HTTPServer) RegisterAPIRoutes(routes ...*APIVersionRouter) {
	for _, router := range routes {
		prefix := "/api/" + string(router.apiVersion)

		h.mux.Handle(
			prefix+"/",
			http.StripPrefix(prefix, router),
		)
	}
}

func (h *HTTPServer) Run(ctx context.Context) error {
	mux := core_http_middleware.ChainMiddleware(h.mux, h.middleware...)
	server := &http.Server{
		Addr:    h.cfg.Addr,
		Handler: mux,
	}
	ch := make(chan error, 1)

	go func() {

		h.log.Warn("Start HTTP Server", slog.String("addr", h.cfg.Addr))
		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("listen and server HTTP: %w", err)
		}
	case <-ctx.Done():
		h.log.Warn("Shutdown HTTP Server")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			h.cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		h.log.Warn("HTTP Server stopped")
	}
	return nil
}
