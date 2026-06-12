package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var fieldAllowSort []string = []string{
	"_id",
	"reference",
	"amount",
	"created_at",
	"updated_at",
}

type Repository interface {
	Create(ctx context.Context, entity *entity.Transaction) error
	UpdateTopUpState(ctx context.Context, refCode string, status constants.TransactionStatus, balanceBefore uint64, balanceAfter uint64) error
	GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error)
	GetAllByAccountNumber(ctx context.Context, accountNumber string, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error)
	GetAllByWalletId(ctx context.Context, walletId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error)
	GetOneByRef(ctx context.Context, refCode string, userId bson.ObjectID, entity *entity.Transaction) error
}

type repository struct {
	collection *mongo.Collection
	refCode    *utils.ReferenceCode
}

func NewRepository(db *mongo.Database, refCode *utils.ReferenceCode) Repository {
	return &repository{
		collection: db.Collection("transactions"),
		refCode:    refCode,
	}
}

func (r *repository) Create(ctx context.Context, entity *entity.Transaction) error {
	if entity.Reference == "" {
		refCode, err := r.refCode.Next(entity.Type.Short())
		if err != nil {
			return err
		}
		entity.Reference = refCode
	}

	entity.CreatedAt = time.Now()
	entity.Currency = "IDR"

	res, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		entity.ID = oid
	}

	return nil
}

func (r *repository) UpdateTopUpState(
	ctx context.Context,
	refCode string,
	status constants.TransactionStatus,
	balanceBefore uint64,
	balanceAfter uint64,
) error {
	filter := bson.M{"reference": refCode}

	update := bson.M{
		"$set": bson.M{
			"status":                     status,
			"destination.balance_before": balanceBefore,
			"destination.balance_after":  balanceAfter,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error) {
	filter := bson.M{}

	userFilter := bson.M{
		"$or": bson.A{
			bson.M{"source.user_id": userId},
			bson.M{"destination.user_id": userId},
		},
	}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		keywordFilter := bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"reference": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"note": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
		filter = bson.M{
			"$and": bson.A{
				userFilter,
				keywordFilter,
			},
		}
	} else {
		filter = userFilter
	}

	res, err := utils.Paginate[entity.Transaction](ctx, r.collection, utils.PaginationParam{
		Page:      q.Page,
		Limit:     q.Limit,
		SortBy:    q.SortBy,
		SortOrder: utils.ValidateSortField(fieldAllowSort, q.SortOrder, "created_at"),
		Filter:    filter,
	})

	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *repository) GetAllByAccountNumber(ctx context.Context, accountNumber string, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error) {
	filter := bson.M{}

	userFilter := bson.M{
		"$or": bson.A{
			bson.M{
				"$and": bson.A{
					bson.M{"source.account_number": accountNumber},
					bson.M{"source.user_id": userId},
				},
			},
			bson.M{
				"$and": bson.A{
					bson.M{"destination.account_number": accountNumber},
					bson.M{"destination.user_id": userId},
				},
			},
		},
	}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		keywordFilter := bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"reference": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"note": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
		filter = bson.M{
			"$and": bson.A{
				userFilter,
				keywordFilter,
			},
		}
	} else {
		filter = userFilter
	}

	res, err := utils.Paginate[entity.Transaction](ctx, r.collection, utils.PaginationParam{
		Page:      q.Page,
		Limit:     q.Limit,
		SortBy:    q.SortBy,
		SortOrder: utils.ValidateSortField(fieldAllowSort, q.SortOrder, "created_at"),
		Filter:    filter,
	})

	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *repository) GetAllByWalletId(ctx context.Context, walletId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.Transaction], error) {
	filter := bson.M{}

	userFilter := bson.M{
		"$or": bson.A{
			bson.M{"source.wallet_id": walletId},
			bson.M{"destination.wallet_id": walletId},
		},
	}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		keywordFilter := bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"reference": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"note": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
		filter = bson.M{
			"$and": bson.A{
				userFilter,
				keywordFilter,
			},
		}
	} else {
		filter = userFilter
	}

	res, err := utils.Paginate[entity.Transaction](ctx, r.collection, utils.PaginationParam{
		Page:      q.Page,
		Limit:     q.Limit,
		SortBy:    q.SortBy,
		SortOrder: utils.ValidateSortField(fieldAllowSort, q.SortOrder, "created_at"),
		Filter:    filter,
	})

	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *repository) GetOneByRef(ctx context.Context, refCode string, userId bson.ObjectID, entity *entity.Transaction) error {
	filter := bson.M{
		"reference": refCode,
		"$or": bson.A{
			bson.M{"source.user_id": userId},
			bson.M{"destination.user_id": userId},
		},
	}
	res := r.collection.FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.ErrDataNotFound
		}
		return err
	}

	if err := res.Decode(entity); err != nil {
		return err
	}
	return nil
}
