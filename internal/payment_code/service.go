package payment_code

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/internal/user"
	"github.com/nrhox/cpay-service/internal/wallet"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service struct {
	walletRepo  wallet.Repository
	userRepo    user.Repository
	paymentRepo Repository
	log         *slog.Logger
	mu          sync.Mutex
}

func NewService(paymentRepo Repository, walletRepo wallet.Repository, userRepo user.Repository, log *slog.Logger) *Service {
	return &Service{
		paymentRepo: paymentRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		log:         log,
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

	return &paymentCode, nil
}
