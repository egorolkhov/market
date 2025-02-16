package middleware

import (
	"avito/internal/models"
	"avito/internal/utils"
	"net/http"
)

func Cookie(h http.HandlerFunc) http.HandlerFunc {
	foo := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			response := models.ErrorResponse{Errors: "unauthorized"}
			utils.JsonResponse(w, http.StatusUnauthorized, response)
			return
		}
		r.Header.Set("Authorization", cookie.Value)

		h.ServeHTTP(w, r)
	}
	return foo
}
