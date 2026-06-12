package middleware

import (
	"context"
	"net/http"

	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthCredential struct {
	Id    bson.ObjectID
	Token string
}

const CREDENTIAL_KEY = "credential"

func (m *Middlware) IsAuth(strict bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, id := security.GetRefreshToken(r)
			if (id == bson.NilObjectID || token == "") && strict {
				security.DeleteAccessToken(w)
				security.DeleteRefreshToken(w)
				response.ParseError(w, errmsg.ErrMissingToken, m.log)
				return
			}

			if id != bson.NilObjectID && token != "" {
				ctx := context.WithValue(r.Context(), CREDENTIAL_KEY, AuthCredential{
					Id:    id,
					Token: token,
				})
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetAuthCredential(c context.Context) (*AuthCredential, error) {
	credential, ok := c.Value(CREDENTIAL_KEY).(AuthCredential)
	if !ok {
		return nil, errmsg.ErrMissingToken
	}

	return &credential, nil
}
