package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const USER_PAYLOAD_KEY = "user_payload"

func (m *Middlware) AccessGuard(strict bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := security.GetAccessToken(r)
			if token == "" {
				tokenAuth := r.Header.Get("Authorization")
				tokenParts := strings.Split(tokenAuth, " ")
				if strict && (tokenAuth == "" || len(tokenParts) != 2 || tokenParts[0] != "Bearer") {
					response.ParseError(w, errmsg.ErrMissingToken, m.log)
					return
				}

				if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
					token = tokenParts[1]
				}
			}

			if strict && (token == "") {
				response.ParseError(w, errmsg.ErrMissingToken, m.log)
				return
			}

			if token != "" {
				payload, err := m.tokenManager.Verify(token)
				if err != nil {
					response.ParseError(w, err, m.log)
					return
				}

				if payload == nil {
					response.ParseError(w, errmsg.ErrMissingToken, m.log)
					return
				}

				ctx := context.WithValue(r.Context(), USER_PAYLOAD_KEY, payload)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

type UserPayload struct {
	UserID bson.ObjectID
	RoleId constants.Role
}

func GetPayloadUser(c context.Context) (*UserPayload, error) {
	payload, ok := c.Value(USER_PAYLOAD_KEY).(*security.AuthPayload)
	if !ok {
		return nil, errmsg.ErrMissingToken
	}

	oid, _ := bson.ObjectIDFromHex(payload.UserID)

	return &UserPayload{
		UserID: oid,
		RoleId: payload.RoleId,
	}, nil
}
