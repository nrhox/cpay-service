package transaction

import (
	"context"
	"log/slog"

	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	transactionRepo Repository
	log             *slog.Logger
}

func NewService(transactionRepo Repository, log *slog.Logger) *Service {
	return &Service{
		transactionRepo: transactionRepo,
		log:             log,
	}
}

func (s *Service) GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) ([]entity.Transaction, response.ResMetaPaginate, error) {
	res, err := s.transactionRepo.GetAllByUserId(ctx, userId, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) GetAllByAccountNumber(ctx context.Context, accountNumber string, userId bson.ObjectID, q utils.QueryParams) ([]entity.Transaction, response.ResMetaPaginate, error) {
	res, err := s.transactionRepo.GetAllByAccountNumber(ctx, accountNumber, userId, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) GetOneByRefAndUserId(ctx context.Context, ref string, userId bson.ObjectID) (*entity.Transaction, error) {
	var transaction entity.Transaction

	if err := s.transactionRepo.GetOneByRef(ctx, ref, userId, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}
