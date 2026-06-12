package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func (m *Middlware) ThrottleLimit() func(http.Handler) http.Handler {
	return middleware.ThrottleWithOpts(middleware.ThrottleOpts{
		Limit:          20,
		BacklogLimit:   50,
		BacklogTimeout: 10 * time.Second,
	})
}
