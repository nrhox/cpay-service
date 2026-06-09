package wallet

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/internal/transaction"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	walletRepo      Repository
	transactionRepo transaction.Repository
	userRepo        user.Repository
	log             *slog.Logger
	mu              sync.Mutex
}

func NewService(walletRepo Repository, transactionRepo transaction.Repository, userRepo user.Repository, log *slog.Logger) *Service {
	return &Service{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		log:             log,
	}
}

func (s *Service) CreateWallet(ctx context.Context, userId bson.ObjectID, dto CreateWallet) (*entity.Wallet, error) {
	availablePrimary, err := s.walletRepo.AvailableWalletPrimary(ctx, userId)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(dto.Pin), 12)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	newWallet := entity.Wallet{
		UserID:    userId,
		Status:    constants.WalletActive,
		Name:      dto.Name,
		IsPrimary: availablePrimary,
		Pin:       string(pinHash),
	}

	if err := s.walletRepo.Create(ctx, &newWallet); err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	newWallet.Pin = ""

	return &newWallet, nil
}

func (s *Service) GetAllWalletByUserID(ctx context.Context, userId bson.ObjectID) ([]entity.Wallet, error) {
	wallets, err := s.walletRepo.GetAllByUserId(ctx, userId)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	return wallets, nil
}

func (s *Service) SetPrimary(ctx context.Context, userId bson.ObjectID, data SetPrimaryWallet) error {
	return s.walletRepo.SetOnePrimary(ctx, userId, data.WalletId)
}

func (s *Service) SetSuspend(ctx context.Context, id bson.ObjectID) error {
	if err := s.userRepo.SetStatus(ctx, id, constants.UserSuspended); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.ErrDataNotFound
		}
		return err
	}

	if err := s.walletRepo.SetAllStatusByUserId(ctx, id, constants.WalletSuspended); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetActive(ctx context.Context, id bson.ObjectID) error {
	if err := s.userRepo.SetStatus(ctx, id, constants.UserActive); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.ErrDataNotFound
		}
		return err
	}

	if err := s.walletRepo.SetAllStatusByUserId(ctx, id, constants.WalletActive); err != nil {
		return err
	}
	return nil
}

func (s *Service) Transfer(ctx context.Context, userId bson.ObjectID, data TransferBalance) (*entity.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var destinationWallet entity.Wallet
	if err := s.walletRepo.FindActiveByAccountNumber(ctx, data.DestionationWallet, &destinationWallet); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrDestionationWalletNotFound
		}
		return nil, err
	}

	var currentWallet entity.Wallet
	if err := s.walletRepo.FindActiveByIdWithUser(ctx, data.WalletId, userId, &currentWallet); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrWalletNotFound
		}
		return nil, err
	}

	if currentWallet.ID == destinationWallet.ID {
		return nil, errmsg.ErrDestionationWalletNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentWallet.Pin), []byte(data.Pin)); err != nil {
		return nil, errmsg.ErrPinNoMatch
	}

	var currentUser entity.User
	if err := s.userRepo.FindUserActiveById(ctx, userId, &currentUser); err != nil {
		return nil, err
	}

	var destionationUser entity.User
	if err := s.userRepo.FindUserActiveById(ctx, destinationWallet.UserID, &destionationUser); err != nil {
		return nil, err
	}

	// update current wallet
	if err := s.walletRepo.UpdateBalance(ctx, currentWallet.ID, -int(data.Amount)); err != nil {
		return nil, err
	}

	// update destionation wallet
	if err := s.walletRepo.UpdateBalance(ctx, destinationWallet.ID, int(data.Amount)); err != nil {
		return nil, err
	}

	currentBalanceAfter := currentWallet.Balance - data.Amount
	destionationBalanceAfter := destinationWallet.Balance + data.Amount

	newTransaction := entity.Transaction{
		Type:   constants.TypeTransfer,
		Amount: data.Amount,
		Status: constants.StatusSuccess,

		Source: &entity.TransactionParty{
			UserID:        currentUser.ID,
			Username:      currentUser.FullName,
			WalletID:      currentWallet.ID,
			WalletName:    currentWallet.Name,
			AccountNumber: currentWallet.AccountNumber,
			BalanceBefore: currentWallet.Balance,
			BalanceAfter:  &currentBalanceAfter,
		},
		Destination: &entity.TransactionParty{
			UserID:        destionationUser.ID,
			Username:      destionationUser.FullName,
			WalletID:      destinationWallet.ID,
			WalletName:    destinationWallet.Name,
			AccountNumber: destinationWallet.AccountNumber,
			BalanceBefore: destinationWallet.Balance,
			BalanceAfter:  &destionationBalanceAfter,
		},
	}

	if err := s.transactionRepo.Create(ctx, &newTransaction); err != nil {
		return nil, err
	}

	return &newTransaction, nil
}

func (s *Service) GetOneByAccountNumber(ctx context.Context, userId *bson.ObjectID, accountNumber string) (*entity.Wallet, error) {
	var wallet entity.Wallet

	if userId != nil {
		if err := s.walletRepo.FindByAccounNumberWithUser(ctx, *userId, accountNumber, &wallet); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errmsg.ErrDataNotFound
			}
			return nil, err
		}
	} else {
		if err := s.walletRepo.FindByAccountNumber(ctx, accountNumber, constants.WalletActive, &wallet); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errmsg.ErrDataNotFound
			}
			return nil, err
		}
	}

	wallet.Pin = ""

	return &wallet, nil
}
