package topup_request

import (
	"context"
	"errors"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/internal/transaction"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service struct {
	topupRepo       Repository
	userRepo        user.Repository
	walletRepo      wallet.Repository
	transactionRepo transaction.Repository
}

func NewService(
	topupRepo Repository,
	userRepo user.Repository,
	walletRepo wallet.Repository,
	transactionRepo transaction.Repository,
) *Service {
	return &Service{
		topupRepo:       topupRepo,
		userRepo:        userRepo,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *Service) CreateRequest(ctx context.Context, userId bson.ObjectID, dto RequestTopup) (*entity.TopupRequest, error) {
	var user entity.User

	if err := s.userRepo.FindUserActiveById(ctx, userId, &user); err != nil {
		return nil, err
	}

	var wallet entity.Wallet

	if err := s.walletRepo.FindByAccounNumber(ctx, userId, dto.WalletNumber, &wallet); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrWalletNotFound
		}
		return nil, err
	}

	newTransaction := entity.Transaction{
		Type:   constants.TypeTopup,
		Title:  "Top up sebesar " + utils.FormatCurrencyByRegion(float64(dto.Amount), "ID"),
		Amount: dto.Amount,
		Status: constants.StatusPending,
		Destination: &entity.TransactionParty{
			UserID:        userId,
			WalletID:      wallet.ID,
			WalletName:    wallet.Name,
			AccountNumber: wallet.AccountNumber,
			BalanceBefore: wallet.Balance,
			BalanceAfter:  nil,
			Username:      user.FullName,
		},
	}

	if err := s.transactionRepo.Create(ctx, &newTransaction); err != nil {
		return nil, err
	}

	newTopUp := entity.TopupRequest{
		UserID:    user.ID,
		WalletID:  wallet.ID,
		Amount:    dto.Amount,
		Reference: newTransaction.Reference,
		Status:    constants.StatusPending,
	}

	if err := s.topupRepo.Create(ctx, &newTopUp); err != nil {
		return nil, err
	}

	return &newTopUp, nil
}
