package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/internal/providers"
	"github.com/nrhox/cpay-service/internal/session"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service struct {
	userSvc    *user.Service
	userRepo   user.Repository
	sessionSvc *session.Service
	WalletSvc  *wallet.Service
	log        *slog.Logger
}

func NewService(
	userSvc *user.Service,
	userRepo user.Repository,
	sessionSvc *session.Service,
	WalletSvc *wallet.Service,
	log *slog.Logger,
) *Service {
	return &Service{
		userSvc:    userSvc,
		userRepo:   userRepo,
		sessionSvc: sessionSvc,
		WalletSvc:  WalletSvc,
		log:        log,
	}
}

func (s *Service) LoginUser(ctx context.Context, data *providers.Profile) (session *entity.Session, isComplate bool, err error) {
	user := user.UserInfo{
		FullName:  data.FullName,
		Email:     data.Email,
		AvatarUrl: data.Picture,
	}
	prov := entity.AuthProvider{
		ID:       data.ProviderID,
		Provider: data.ProviderName,
	}

	newUser, err := s.userSvc.Upsert(ctx, user, prov)
	if err != nil {
		if errors.Is(err, errmsg.ErrAccountSuspend) {
			return nil, false, err
		}
		s.log.Error(err.Error())
		return nil, false, errmsg.ErrOauthAuthProcessFailed
	}

	session, err = s.sessionSvc.Create(ctx, newUser.ID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, false, errmsg.ErrOauthAuthProcessFailed
	}

	return session, newUser.Status != constants.UserUncomplateRegister, nil
}

func (s *Service) RefreshToken(ctx context.Context, tokenId bson.ObjectID, token string) (*entity.User, error) {
	session, err := s.sessionSvc.GetAvailable(ctx, tokenId, token)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, err
	}

	user, err := s.userSvc.GetOneNoSuspend(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Logout(ctx context.Context, tokenId bson.ObjectID, token string) error {
	return s.sessionSvc.Delete(ctx, tokenId, token)
}

func (s *Service) IncomplateRegister(ctx context.Context, tokenId bson.ObjectID, token string, dto wallet.CreateWallet) error {
	session, err := s.sessionSvc.GetAvailable(ctx, tokenId, token)
	if err != nil {
		return err
	}

	if session == nil {
		return errmsg.ErrMissingToken
	}

	isIncomplate, err := s.userRepo.CheckUserStatus(ctx, session.UserID, constants.UserUncomplateRegister)
	if err != nil {
		return err
	}

	if !isIncomplate {
		return nil
	}

	if _, err := s.WalletSvc.CreateWallet(ctx, session.UserID, dto); err != nil {
		return err
	}

	if err := s.userRepo.SetStatus(ctx, session.UserID, constants.UserActive); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.ErrDataNotFound
		}
		return err
	}

	return nil
}
