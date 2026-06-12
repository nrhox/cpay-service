package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	userRepo Repository
	log      *slog.Logger
}

func NewService(userRepo Repository, log *slog.Logger) *Service {
	return &Service{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *Service) Upsert(ctx context.Context, info UserInfo, provider entity.AuthProvider) (*entity.User, error) {
	var user entity.User

	if err := s.userRepo.GetOneByEmail(ctx, info.Email, &user); err != nil {
		if !errors.Is(err, errmsg.ErrUserNotFound) {
			return nil, err
		}
	}

	if !user.ID.IsZero() {
		isInsert := false
		for _, prov := range user.OAuthProviders {
			if prov.Provider != provider.Provider {
				user.OAuthProviders = append(user.OAuthProviders, provider)
				isInsert = true
				break
			}
		}
		if isInsert {
			if err := s.userRepo.UpsertProvider(ctx, user.ID, provider); err != nil {
				return nil, err
			}
		}
	}

	if user.ID.IsZero() {
		user.AvatarUrl = info.AvatarUrl
		user.Email = info.Email
		user.FullName = info.FullName
		user.RoleID = constants.RoleUser
		user.OAuthProviders = []entity.AuthProvider{provider}

		if err := s.userRepo.NewUser(ctx, &user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *Service) GetOneNoSuspend(ctx context.Context, id bson.ObjectID) (*entity.User, error) {
	var user entity.User

	if err := s.userRepo.GetOneNoSuspendById(ctx, id, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) GetAll(ctx context.Context, notId bson.ObjectID, q utils.QueryParams) ([]entity.User, response.ResMetaPaginate, error) {
	res, err := s.userRepo.GetAll(ctx, notId, q)
	if err != nil {
		return nil, response.ResMetaPaginate{}, err
	}

	return res.Data, response.ResMetaPaginate{
		TotalPage: res.TotalPage,
		TotalData: res.TotalData,
	}, nil
}

func (s *Service) GetOne(ctx context.Context, id bson.ObjectID) (*entity.User, error) {
	var user entity.User

	if err := s.userRepo.GetOneById(ctx, id, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
