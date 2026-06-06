package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/auth"
	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/internal/delivery/route"
	"github.com/nrhox/cpay-service/internal/providers"
	"github.com/nrhox/cpay-service/internal/session"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Bootstrap struct {
	Route  *chi.Mux
	DB     *mongo.Database
	Logger *slog.Logger
	Cfg    *config.Config
}

func (b *Bootstrap) Init() {
	tokenManager := security.NewTokenManager(
		b.Cfg.Session.JwtPrivateKey,
		b.Cfg.Session.JwtPublicKey,
	)

	providers.NewGitHubProvider(b.Cfg.Providers.Github)

	userRepo := user.NewRepository(b.DB)
	sessionRepo := session.NewRepository(b.DB)

	userService := user.NewService(userRepo, b.Logger)
	sessionService := session.NewService(b.Cfg.Session, sessionRepo, b.Logger)
	authService := auth.NewService(userService, sessionService, b.Logger)

	authHandler := auth.NewHandler(authService, b.Logger, &b.Cfg.Session, b.Cfg.FrontendUrl, tokenManager)

	middleware := middleware.NewMiddlware(tokenManager, b.Logger, b.Cfg)

	route.NewRoute(b.Route, authHandler, middleware)
}

func (b *Bootstrap) PrintAllRoute() {
	if b.Cfg.Mode == config.MODE_DEBUG {
		chi.Walk(b.Route, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			funcName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			lastDot := strings.LastIndex(funcName, ".")

			if lastDot != -1 {
				funcName = funcName[lastDot+1:]
			}

			fmt.Printf("[%s]\t %s -> %s\n", method, route, funcName)
			return nil
		})
	}
}
