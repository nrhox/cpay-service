package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nrhox/cpay-service/internal/app"
	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/pkg/database"
)

func main() {
	cfg := config.Load()

	dbMongo, disconnent, err := database.NewMongoConnetion(*cfg)
	if err != nil {
		panic(err)
	}
	defer disconnent(context.Background())

	log := config.LoadLogger(cfg)
	log.Info(
		"application run",
		slog.Attr{
			Key:   "pid",
			Value: slog.IntValue(os.Getpid()),
		},
		slog.Attr{
			Key:   "os",
			Value: slog.StringValue(runtime.GOOS),
		},
		slog.Attr{
			Key:   "arch",
			Value: slog.StringValue(runtime.GOARCH),
		},
	)

	r := chi.NewRouter()
	r.Use(setupCORS(cfg))
	r.Use(middleware.Recoverer)
	if cfg.Mode == config.MODE_DEBUG {
		r.Use(middleware.Logger)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	appBootstrap := app.Bootstrap{
		Route:  r,
		DB:     dbMongo,
		Logger: log,
		Cfg:    cfg,
	}
	appBootstrap.Init()
	appBootstrap.PrintAllRoute()

	http.ListenAndServe(cfg.AppPort, r)
}

func setupCORS(cfg *config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
