package wallet

import (
	"context"
	"log/slog"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	walletRepo Repository
	log        *slog.Logger
}

func NewService(walletRepo Repository, log *slog.Logger) *Service {
	return &Service{
		walletRepo: walletRepo,
		log:        log,
	}
}

func (s *Service) CreateWallet(ctx context.Context, userId bson.ObjectID, dto CreateWallet) (*entity.Wallet, error) {
	availablePrimary, err := s.walletRepo.AvailableWalletPrimary(ctx, userId)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	newWallet := entity.Wallet{
		UserID:    userId,
		Status:    constants.WalletActive,
		Name:      dto.Name,
		IsPrimary: availablePrimary,
	}

	if err := s.walletRepo.Create(ctx, &newWallet); err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	return &newWallet, nil
}
