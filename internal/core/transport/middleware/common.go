package core_http_middleware

import (
	"context"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const requestIdHeader = "X-Request-ID"

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}
			requestId := r.Header.Get(requestIdHeader)
			if requestId == "" {
				requestId = uuid.NewString()
			}

			r.Header.Set(requestIdHeader, requestId)
			w.Header().Set(requestIdHeader, requestId)
			next.ServeHTTP(w, r)
		})
	}
}

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}
			requestId := r.Header.Get(requestIdHeader)

			l := log.With(
				slog.String("request_id", requestId),
				slog.String("url", r.URL.String()),
			)

			ctx := context.WithValue(r.Context(), "log", l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(log, w)
			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(p, "during handle HTTP request got unexpected panic")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}
			rw := core_http_response.NewResponseWriter(w)

			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			before := time.Now().UTC()
			log.Info(
				">>>> INCOMING HTTP REQUEST",
				slog.String("Method", r.Method),
				slog.Time("time", before),
			)

			next.ServeHTTP(rw, r)

			log.Info(
				"<<<< DONE HTTP REQUEST",
				slog.Duration("duration", time.Since(before)),
				slog.Int("statusCode", rw.GetStatusCodeOrPanic()),
			)
		})
	}
}

func Cors() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Вместо "*" берем Origin из заголовков самого запроса
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// 2. КРИТИЧЕСКИ ВАЖНО: разрешаем передачу кук/credentials
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")

			// Обработка preflight-запросов
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
