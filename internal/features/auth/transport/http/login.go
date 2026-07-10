package auth_transport_http

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := getLogger(r.Context())

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("decode login request", "error", err)
		http.Error(w, `{"error":"некорректный запрос"}`, http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Warn("login failed", "error", err)
		http.Error(w, `{"error":"неверное имя пользователя или пароль"}`, http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(sessionCookieTTL),
	})

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
