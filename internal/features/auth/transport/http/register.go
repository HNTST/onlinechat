package auth_transport_http

import (
	core_http_request "discord/internal/core/transport/request"
	"encoding/json"
	"net/http"
	"time"
)

// Register creates a new user and auto-authenticates
//
//	@Summary		Регистрация пользователя
//	@Description	Создание новой учётной записи с авто-логином (устанавливается session_token)
//	@Tags			v2 / Аутентификация
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Данные для регистрации"
//	@Success		201		{object}	domain.User
//	@Failure		400		{object}	object{}
//	@Router			/api/v2/auth/register [post]
func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := getLogger(r.Context())

	var req RegisterRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		log.Warn("decode register request", "error", err)
		http.Error(w, `{"error":"некорректный запрос"}`, http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Warn("register failed", "error", err)
		http.Error(w, `{"error":"регистрация не удалась"}`, http.StatusBadRequest)
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
