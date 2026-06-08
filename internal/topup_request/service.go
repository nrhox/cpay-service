package topup_request

import (
	"context"
	"errors"
	"sync"

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
)

type Service struct {
	topupRepo       Repository
	userRepo        user.Repository
	walletRepo      wallet.Repository
	transactionRepo transaction.Repository
	mu              sync.Mutex
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

	if err := s.walletRepo.FindByAccounNumberWithUser(ctx, userId, dto.WalletNumber, &wallet); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errmsg.ErrWalletNotFound
		}
		return nil, err
	}

	newTransaction := entity.Transaction{
		Type:   constants.TypeTopup,
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

func (s *Service) GetAll(ctx context.Context, q utils.QueryParams) ([]entity.TopupRequest, response.ResMetaPaginate, error) {
	res, err := s.topupRepo.GetAll(ctx, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) GetOneById(ctx context.Context, id bson.ObjectID) (*entity.TopupRequest, error) {
	var topUp entity.TopupRequest
	if err := s.topupRepo.GetOneById(ctx, id, &topUp); err != nil {
		return nil, err
	}

	return &topUp, nil
}

func (s *Service) SetApproved(ctx context.Context, id bson.ObjectID) (*entity.TopupRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var topUp entity.TopupRequest
	if err := s.topupRepo.GetOneById(ctx, id, &topUp); err != nil {
		return nil, err
	}

	var wallet entity.Wallet
	if err := s.walletRepo.FindById(ctx, topUp.WalletID, &wallet); err != nil {
		return nil, err
	}

	if wallet.Status != constants.WalletActive {
		return nil, errmsg.ErrWalletNotFound
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, int(topUp.Amount)); err != nil {
		return nil, err
	}

	if err := s.transactionRepo.UpdateTopUpState(
		ctx,
		topUp.Reference,
		constants.StatusSuccess,
		wallet.Balance,
		wallet.Balance+topUp.Amount,
	); err != nil {
		return nil, err
	}

	if err := s.topupRepo.SetStatus(ctx, id, constants.StatusSuccess); err != nil {
		return nil, err
	}

	topUp.Status = constants.StatusSuccess

	return &topUp, nil
}

func (s *Service) SetReject(ctx context.Context, id bson.ObjectID) (*entity.TopupRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var topUp entity.TopupRequest
	if err := s.topupRepo.GetOneById(ctx, id, &topUp); err != nil {
		return nil, err
	}

	var wallet entity.Wallet
	if err := s.walletRepo.FindById(ctx, topUp.WalletID, &wallet); err != nil {
		return nil, err
	}

	if wallet.Status != constants.WalletActive {
		return nil, errmsg.ErrWalletNotFound
	}

	if err := s.transactionRepo.UpdateTopUpState(
		ctx,
		topUp.Reference,
		constants.StatusCancelled,
		wallet.Balance,
		wallet.Balance,
	); err != nil {
		return nil, err
	}

	if err := s.topupRepo.SetStatus(ctx, id, constants.StatusCancelled); err != nil {
		return nil, err
	}

	topUp.Status = constants.StatusCancelled

	return &topUp, nil
}
