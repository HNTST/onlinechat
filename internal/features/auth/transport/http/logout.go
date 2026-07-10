package auth_transport_http

import (
	"net/http"
)

// Logout clears session and cookie
//
//	@Summary		Выход из системы
//	@Description	Завершение сессии, удаление session_token cookie
//	@Tags			v2 / Аутентификация
//	@Produce		json
//	@Success		200		{object}	object{}
//	@Router			/api/v2/auth/logout [post]
func (h *AuthHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	log := getLogger(r.Context())

	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
		log.Warn("logout failed", "error", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}
