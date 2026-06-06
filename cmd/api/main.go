package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/go-chi/chi/v5"
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
