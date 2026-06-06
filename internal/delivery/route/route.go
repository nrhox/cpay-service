package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
)

func NewRoute(
	r *chi.Mux,
	authH *auth.Handler,
	m *middleware.Middlware,
) {
	r.Route("/api/auth", func(r chi.Router) {
		r.With(m.GuestOnly).Get("/{provider}", authH.Login)
		r.With(m.GuestOnly, m.InjectUser).Get("/{provider}/callback", authH.Callback)
		r.With(m.IsAuth(true)).Post("/incomplate", authH.IncomplateRegister)
		r.With(m.IsAuth(true)).Get("/__refresh", authH.RefreshToken)
		r.With(m.IsAuth(false)).Get("/logout", authH.Logout)
	})
}
