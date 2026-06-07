package topup_request

import (
	"context"
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var fieldAllowSort []string = []string{
	"wallet_id",
	"amount",
	"reference",
	"status",
	"requested_at",
}

type Repository interface {
	Create(ctx context.Context, entity *entity.TopupRequest) error
	GetAll(ctx context.Context, q utils.QueryParams) (utils.PaginationResult[entity.TopupRequest], error)
}

type repository struct {
	collection *mongo.Collection
	refCode    *utils.ReferenceCode
}

func NewRepository(db *mongo.Database, refCode *utils.ReferenceCode) Repository {
	return &repository{
		collection: db.Collection("topup_request"),
		refCode:    refCode,
	}
}

func (r *repository) Create(ctx context.Context, entity *entity.TopupRequest) error {
	entity.RequestedAt = time.Now()
	if entity.Reference == "" {
		refCode, err := r.refCode.Next(constants.TypeTopup.Short())
		if err != nil {
			return err
		}

		entity.Reference = refCode
	}

	res, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		entity.ID = oid
	}

	return nil
}

func (r *repository) GetAll(ctx context.Context, q utils.QueryParams) (utils.PaginationResult[entity.TopupRequest], error) {
	filter := bson.M{}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		filter = bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"wallet_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"amount": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"reference": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"status": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
	}

	res, err := utils.Paginate[entity.TopupRequest](ctx, r.collection, utils.PaginationParam{
		Page:      q.Page,
		Limit:     q.Limit,
		SortBy:    q.SortBy,
		SortOrder: utils.ValidateSortField(fieldAllowSort, q.SortOrder, "requested_at"),
		Filter:    filter,
	})

	if err != nil {
		return res, err
	}

	return res, nil
}
