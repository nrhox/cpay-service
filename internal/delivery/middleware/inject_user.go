package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/providers"
	"github.com/nrhox/cpay-service/pkg/errmsg"
)

const USER_INJECT_KEY = "user_inject"

func (m *Middlware) InjectUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		injectProvider := chi.URLParam(r, "provider") == "inject"
		if injectProvider && m.config.Mode == config.MODE_DEBUG && m.config.UserMock.WithMock && m.config.UserMock.User != nil {
			user := m.config.UserMock.User
			mockAuth := providers.Profile{
				Email:        user.Email,
				FullName:     user.FullName,
				Picture:      user.AvatarUrl,
				ProviderName: user.OAuthProviders[0].Provider,
				ProviderID:   user.OAuthProviders[0].ID,
			}

			ctx := context.WithValue(r.Context(), USER_INJECT_KEY, &mockAuth)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserInjection(c context.Context) (*providers.Profile, error) {
	payload, ok := c.Value(USER_INJECT_KEY).(*providers.Profile)
	if !ok {
		return nil, errmsg.ErrDataNotFound
	}

	return payload, nil
}
