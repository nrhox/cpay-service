package middleware

import (
	"net/http"

	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m *Middlware) GuestOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, id := security.GetRefreshToken(r)
		if token != "" || id != bson.NilObjectID {
			response.Json(w, http.StatusOK, response.ResJson{
				Message: "You are already logged in",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
