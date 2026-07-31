package auth_transport_http

import (
	"discord/internal/core/domain"
	"encoding/json"
	"net/http"
	"time"
)

// Login authenticates user and sets session cookie
//
//	@Summary		Аутентификация пользователя
//	@Description	Вход по username/password, устанавливает HttpOnly cookie session_token
//	@Tags			v2 / Аутентификация
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AuthRequest	true	"Данные для входа"
//	@Success		200		{object}	domain.User
//	@Failure		401		{object}	object{}
//	@Router			/api/v2/auth/login [post]
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
		Expires:  time.Now().UTC().Add(domain.SessionTTL),
	})

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
