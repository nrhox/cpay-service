package payment_code

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/internal/transaction"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	walletRepo      wallet.Repository
	userRepo        user.Repository
	paymentRepo     Repository
	transactionRepo transaction.Repository
	log             *slog.Logger
	mu              sync.Mutex
}

func NewService(paymentRepo Repository, walletRepo wallet.Repository, userRepo user.Repository, transactionRepo transaction.Repository, log *slog.Logger) *Service {
	return &Service{
		paymentRepo:     paymentRepo,
		walletRepo:      walletRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		log:             log,
	}
}

func (s *Service) CreatePaymentCode(ctx context.Context, userId bson.ObjectID, data CreatePaymentCode) (*entity.PaymentCode, error) {
	var currentWallet entity.Wallet
	if err := s.walletRepo.FindActiveByIdWithUser(ctx, data.WalletId, userId, &currentWallet); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrWalletNotFound
		}
		return nil, err
	}

	var currentUser entity.User
	if err := s.userRepo.FindUserActiveById(ctx, userId, &currentUser); err != nil {
		return nil, err
	}

	newCode := entity.PaymentCode{
		UserID:   userId,
		WalletID: currentUser.ID,
		Merchant: currentUser.FullName + " - " + currentWallet.Name,
		Amount:   data.Amount,
		Note:     data.Note,
		Status:   constants.PaymentActive,
	}

	if err := s.paymentRepo.Create(ctx, &newCode); err != nil {
		return nil, err
	}

	return &newCode, nil
}

func (s *Service) GetAll(ctx context.Context, q utils.QueryParams) ([]entity.PaymentCode, response.ResMetaPaginate, error) {
	res, err := s.paymentRepo.GetAll(ctx, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) ([]entity.PaymentCode, response.ResMetaPaginate, error) {
	res, err := s.paymentRepo.GetAllByUserId(ctx, userId, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) FindById(ctx context.Context, id bson.ObjectID) (*entity.PaymentCode, error) {
	var paymentCode entity.PaymentCode
	if err := s.paymentRepo.FindById(ctx, id, &paymentCode); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrDataNotFound
		}
		return nil, err
	}

	if paymentCode.Status == constants.PaymentActive && paymentCode.ExpiresAt.Before(time.Now()) {
		if err := s.paymentRepo.SetStatus(ctx, paymentCode.ID, constants.PaymentExpired); err != nil {
			return nil, err
		}

		paymentCode.Status = constants.PaymentExpired
	}

	return &paymentCode, nil
}

func (s *Service) FindByCode(ctx context.Context, code string) (*entity.PaymentCode, error) {
	var paymentCode entity.PaymentCode
	if err := s.paymentRepo.FindByCode(ctx, code, &paymentCode); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrDataNotFound
		}
		return nil, err
	}

	if paymentCode.Status == constants.PaymentActive && paymentCode.ExpiresAt.Before(time.Now()) {
		if err := s.paymentRepo.SetStatus(ctx, paymentCode.ID, constants.PaymentExpired); err != nil {
			return nil, err
		}

		paymentCode.Status = constants.PaymentExpired
	}

	return &paymentCode, nil
}

func (s *Service) SetCancelByAdmin(ctx context.Context, id bson.ObjectID) error {
	if err := s.paymentRepo.SetStatus(ctx, id, constants.PaymentCancelled); err != nil {
		return err
	}

	return nil
}

func (s *Service) SetCancelByUser(ctx context.Context, userId bson.ObjectID, code string) error {
	if err := s.paymentRepo.SetStatusByUserId(ctx, userId, code, constants.PaymentCancelled); err != nil {
		return err
	}

	return nil
}

func (s *Service) PayingCode(ctx context.Context, userId bson.ObjectID, data CreetePayingTransaction) (*entity.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var payCode entity.PaymentCode
	if err := s.paymentRepo.FindByCode(ctx, data.PaymentCode, &payCode); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrPaymentCodeNotFound
		}
		return nil, err
	}

	if payCode.Status != constants.PaymentActive || payCode.ExpiresAt.Before(time.Now()) {
		if err := s.paymentRepo.SetStatus(ctx, payCode.ID, constants.PaymentExpired); err != nil {
			return nil, err
		}

		return nil, errmsg.ErrPaymentCodeNotFound
	}

	if payCode.WalletID == data.WalletId {
		return nil, errmsg.ErrDestionationWalletNotFound
	}

	var destinationWallet entity.Wallet
	if err := s.walletRepo.FindById(ctx, payCode.WalletID, &destinationWallet); err != nil {
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
	if err := s.walletRepo.UpdateBalance(ctx, currentWallet.ID, -int(payCode.Amount)); err != nil {
		return nil, err
	}

	// update destionation wallet
	if err := s.walletRepo.UpdateBalance(ctx, destinationWallet.ID, int(payCode.Amount)); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.SetStatus(ctx, payCode.ID, constants.PaymentPaid); err != nil {
		return nil, err
	}

	currentBalanceAfter := currentWallet.Balance - payCode.Amount
	destionationBalanceAfter := destinationWallet.Balance + payCode.Amount

	newTransaction := entity.Transaction{
		Type:   constants.TypeTransfer,
		Amount: payCode.Amount,
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
