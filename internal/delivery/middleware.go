package delivery

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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
func (sv *OLD_SERVER_HANDLERS) sendError(w http.ResponseWriter, msg string, code int, err error) {
	if rw, ok := w.(*responseWriter); ok {
		rw.setError(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (sv *OLD_SERVER_HANDLERS) MWChain(next http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

func (sv *OLD_SERVER_HANDLERS) CORSMV(next http.Handler) http.Handler {
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
func (sv *OLD_SERVER_HANDLERS) LoggingMV(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// duration := time.Since(start)
		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
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
