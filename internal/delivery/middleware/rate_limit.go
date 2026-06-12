package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

func trustedProxyKeyFunc(r *http.Request) (string, error) {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip), nil
	}
	return httprate.KeyByIP(r)
}

func (m *Middlware) RateLimit(limit int) func(http.Handler) http.Handler {
	return httprate.Limit(
		limit,
		1*time.Minute,
		httprate.WithKeyFuncs(trustedProxyKeyFunc),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, m.config.FrontendUrl, http.StatusFound)
		}),
	)
}
