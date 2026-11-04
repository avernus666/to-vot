package httpserver

import (
	"net/http"
	"strings"

	"github.com/avernus666/averna/apps/api/internal/auth"
)

func requireAuth(service *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "A bearer access token is required")
			return
		}
		user, err := service.AuthenticateAccess(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "The access token is invalid or expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
	})
}
