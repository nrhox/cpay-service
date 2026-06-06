package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
)

func NewRoute(
	r *chi.Mux,
	authH *auth.Handler,
) {
	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/{provider}", authH.Login)
		r.Get("/{provider}/callback", authH.Callback)
	})
}
