package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	sessionRepo Repository
	sessionCfg  config.Session
	log         *slog.Logger
}

func NewService(sessionConfig config.Session, sessionRepo Repository, log *slog.Logger) *Service {
	return &Service{
		sessionRepo: sessionRepo,
		sessionCfg:  sessionConfig,
		log:         log,
	}
}

func (s *Service) Create(ctx context.Context, userId bson.ObjectID) (*entity.Session, error) {
	var token string

	countLoop := 0
	for {
		if countLoop == 5 {
			return nil, errmsg.ErrInfiniteLoop
		}

		t, err := security.GenerateRandomToken(64)
		if err != nil {
			return nil, err
		}

		if err := s.sessionRepo.IsTokenAlready(ctx, security.HashTokenForStorage(t)); err != nil {
			if !errors.Is(err, errmsg.ErrTokenAlreadyExists) && countLoop == 4 {
				return nil, err
			}
		} else {
			token = t
			break
		}

		countLoop++
	}

	newSession := entity.Session{
		UserID:    userId,
		Token:     security.HashTokenForStorage(token),
		ExpiredAt: time.Now().Add(s.sessionCfg.RefreshDuration),
	}

	if err := s.sessionRepo.Create(ctx, &newSession); err != nil {
		return nil, err
	}

	newSession.Token = token
	return &newSession, nil
}

func (s *Service) GetAvailable(ctx context.Context, tokenId bson.ObjectID, token string) (*entity.Session, error) {
	var session entity.Session
	if err := s.sessionRepo.GetValidToken(ctx, &session, tokenId, token); err != nil {
		if errors.Is(err, errmsg.ErrDataNotFound) {
			return nil, err
		}
		s.log.Error(err.Error())
		return nil, errmsg.ErrInternalServer
	}

	return &session, nil
}

func (s *Service) Delete(ctx context.Context, tokenId bson.ObjectID, token string) error {
	if err := s.sessionRepo.Delete(ctx, tokenId, token); err != nil {
		s.log.Error(err.Error())
		return errmsg.ErrInternalServer
	}
	return nil
}
