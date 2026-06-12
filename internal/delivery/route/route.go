package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/internal/payment_code"
	"github.com/nrhox/cpay-service/internal/topup_request"
	"github.com/nrhox/cpay-service/internal/transaction"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
)

const (
	NORMAL_LIMIT    = 100
	SEMI_HARD_LIMIT = 20
	HARD_LIMIT      = 5
)

func NewRoute(
	r *chi.Mux,
	authH *auth.Handler,
	userH *user.Handler,
	topUpH *topup_request.Handler,
	walletH *wallet.Handler,
	paymentCode *payment_code.Handler,
	transactionH *transaction.Handler,
	m *middleware.Middlware,
) {
	r.Route("/api/auth", func(r chi.Router) {
		r.With(m.RateLimit(HARD_LIMIT), m.GuestOnly).Get("/{provider}", authH.Login)
		r.With(m.RateLimit(HARD_LIMIT), m.InjectUser).Get("/{provider}/callback", authH.Callback)
		r.With(m.RateLimit(SEMI_HARD_LIMIT), m.IsAuth(true)).Post("/incomplate", authH.IncomplateRegister)
		r.With(m.RateLimit(SEMI_HARD_LIMIT), m.IsAuth(true)).Get("/__refresh", authH.RefreshToken)
		r.With(m.RateLimit(SEMI_HARD_LIMIT), m.IsAuth(false)).Get("/logout", authH.Logout)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(m.AccessGuard(true))
		r.Route("/admin", func(r chi.Router) {
			r.Use(m.RateLimit(300), m.RoleFlag(constants.RoleAdmin))
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

		r.Group(func(r chi.Router) {
			r.Use(m.RateLimit(NORMAL_LIMIT), m.AccessGuard(true))

			r.Route("/wallet", func(r chi.Router) {
				r.Post("/", walletH.NewWallet)
				r.Get("/", walletH.GetMyWallet)
				r.Get("/{account_number}", walletH.GetWalletByAccountNumber)
				r.Put("/", walletH.SetPrimaryWallet)
			})

			r.Route("/payment", func(r chi.Router) {
				r.Get("/", paymentCode.GetAllMyCode)
				r.Get("/{code}", paymentCode.FindByCode)
				r.Delete("/{code}/cancel", paymentCode.SetCancelByUser)

				r.With(m.RateLimit(HARD_LIMIT), m.ThrottleLimit()).Post("/", paymentCode.PayingCode)
				r.With(m.RateLimit(HARD_LIMIT), m.ThrottleLimit()).Post("/create", paymentCode.CreatePaymentCode)
			})

			r.Route("/transaction", func(r chi.Router) {
				r.Get("/{ref_code}", transactionH.GetOneByRefCurrentUser)
				r.Get("/", transactionH.GetMyTransaction)
				r.Get("/wallet/{account_number}", transactionH.GetMyTransactionByAccountNumber)
			})

			r.Get("/me", userH.GetMe)
			r.With(m.RateLimit(SEMI_HARD_LIMIT)).Post("/top-up", topUpH.RequestTopup)
			r.With(m.RateLimit(HARD_LIMIT), m.ThrottleLimit()).Post("/transfer", walletH.TransferBalance)
		})
	})
}
