package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/internal/payment_code"
	"github.com/nrhox/cpay-service/internal/topup_request"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
)

func NewRoute(
	r *chi.Mux,
	authH *auth.Handler,
	userH *user.Handler,
	topUpH *topup_request.Handler,
	walletH *wallet.Handler,
	paymentCode *payment_code.Handler,
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
				r.Get("/{id}/wallet", walletH.GetWalletUser)
				r.Put("/{id}/suspend", walletH.SetSuspendUser)
				r.Put("/{id}/active", walletH.SetActiveUser)
			})

			r.Route("/top-up", func(r chi.Router) {
				r.Get("/", topUpH.GetAllTopUp)
				r.Get("/{id}", topUpH.GetOneById)
				r.Put("/{id}/approved", topUpH.SetApproved)
				r.Put("/{id}/reject", topUpH.SetReject)
			})

			r.Route("/payment-code", func(r chi.Router) {
				r.Get("/", paymentCode.GetAll)
				r.Get("/{id}", paymentCode.FindById)
				r.Get("/user/{id}", paymentCode.GetAllByUserId)
				r.Delete("/{id}/cancel", paymentCode.SetCancelByAdmin)
			})
		})

		r.Route("/wallet", func(r chi.Router) {
			r.Post("/", walletH.NewWallet)
			r.Get("/", walletH.GetMyWallet)
			r.Get("/{account_number}", walletH.GetWalletByAccountNumber)
			r.Put("/", walletH.SetPrimaryWallet)
		})

		r.Route("/payment", func(r chi.Router) {
			r.Get("/", paymentCode.GetAllMyCode)
			r.Post("/create", paymentCode.CreatePaymentCode)
			r.Get("/{code}", paymentCode.FindByCode)
			r.Delete("/{code}/cancel", paymentCode.SetCancelByUser)
		})

		r.Get("/me", userH.GetMe)
		r.Post("/top-up", topUpH.RequestTopup)
		r.Post("/transfer", walletH.TransferBalance)
	})
}
