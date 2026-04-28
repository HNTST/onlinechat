package delivery

import (
	"encoding/json"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	errDetail  string
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) setError(err error) {
	if err != nil {
		rw.errDetail = err.Error()
	}
}
func (sv *Server) sendError(w http.ResponseWriter, msg string, code int, err error) {
	// Пытаемся достучаться до нашей обертки
	if rw, ok := w.(*responseWriter); ok {
		rw.setError(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (sv *Server) MWChain(next http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

func (sv *Server) CORSMV(next http.Handler) http.Handler {
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
func (sv *Server) LoggingMV(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration.String(),
			"ip", r.RemoteAddr,
		}
		if wrapped.errDetail != "" {
			args = append(args, "err_detail", wrapped.errDetail)
		}
		if wrapped.statusCode >= 500 {
			sv.logger.Error("server error", args...)
		} else if wrapped.statusCode >= 400 {
			sv.logger.Warn("client error", args...)
		} else {
			sv.logger.Info("request processed", args...)
		}
	})
}
