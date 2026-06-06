package middleware

import (
	"net/http"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
)

func (m *Middlware) RoleFlag(flag constants.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth, err := GetPayloadUser(r.Context())
			if err != nil {
				response.ParseError(w, errmsg.ErrMissingToken, m.log)
				return
			}

			if auth.RoleId == flag {
				response.ParseError(w, errmsg.ErrMissingToken, m.log)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
