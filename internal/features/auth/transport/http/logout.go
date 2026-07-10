package auth_transport_http

import (
	"net/http"
)

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
