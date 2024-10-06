package api

import (
	"net/http"

	"github.com/gorilla/csrf"
)

func (api *Api) handleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	encodeJson(w, r, http.StatusOK, map[string]string{"csrf_token": token})
}

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Session.Exists(r.Context(), "authenticatedUserId") {
			_ = encodeJson(w, r, http.StatusUnauthorized, map[string]any{"message": "must be logged in"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
