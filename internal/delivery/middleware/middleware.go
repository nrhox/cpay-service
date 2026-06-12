package middleware

import (
	"log/slog"

	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/pkg/security"
)

type Middlware struct {
	tokenManager *security.TokenManager
	log          *slog.Logger
	config       *config.Config
}

func NewMiddlware(
	tokenManager *security.TokenManager,
	log *slog.Logger,
	config *config.Config,
) *Middlware {
	return &Middlware{
		tokenManager: tokenManager,
		log:          log,
		config:       config,
	}
}
