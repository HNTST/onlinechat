package core_http_middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var authExemptPrefixes = []string{
	"/swagger/",
	"/api/v2/auth/login",
}

func Auth(sessionValidator func(ctx context.Context, token string) (int64, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, prefix := range authExemptPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			cookie, err := r.Cookie("session_token")
			if err != nil {
				writeUnauthorized(w, "требуется авторизация")
				return
			}

			userID, err := sessionValidator(r.Context(), cookie.Value)
			if err != nil {
				writeUnauthorized(w, "недействительная сессия")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
