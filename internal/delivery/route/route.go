package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/internal/topup_request"
	"github.com/nrhox/cpay-service/internal/user"
)

func NewRoute(
	r *chi.Mux,
	authH *auth.Handler,
	userH *user.Handler,
	topUpH *topup_request.Handler,
	m *middleware.Middlware,
) {
	r.Route("/api/auth", func(r chi.Router) {
		r.With(m.GuestOnly).Get("/{provider}", authH.Login)
		r.With(m.GuestOnly, m.InjectUser).Get("/{provider}/callback", authH.Callback)
		r.With(m.IsAuth(true)).Post("/incomplate", authH.IncomplateRegister)
		r.With(m.IsAuth(true)).Get("/__refresh", authH.RefreshToken)
		r.With(m.IsAuth(false)).Get("/logout", authH.Logout)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(m.AccessGuard(true))
		r.Route("/admin", func(r chi.Router) {
			r.Use(m.RoleFlag(constants.RoleAdmin))
			r.Route("/user", func(r chi.Router) {
				r.Get("/", userH.GetAllUser)
				r.Get("/{id}", userH.GetOne)
				r.Put("/{id}/suspend", userH.SetSuspendUser)
				r.Put("/{id}/active", userH.SetActiveUser)
			})

			r.Route("/top-up", func(r chi.Router) {
				r.Get("/", topUpH.GetAllTopUp)
				r.Get("/{id}", topUpH.GetOneById)
				r.Put("/{id}/approved", topUpH.SetApproved)
				r.Put("/{id}/reject", topUpH.SetReject)
			})
		})

		r.Get("/me", userH.GetMe)

		r.Post("/top-up", topUpH.RequestTopup)
	})
}
