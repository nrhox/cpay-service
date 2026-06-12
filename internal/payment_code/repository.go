package payment_code

import (
	"context"
	"time"

	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var fieldAllowSort []string = []string{
	"merchant",
	"code",
	"amount",
	"status",
	"expires_at",
	"created_at",
}

type Repository interface {
	Create(ctx context.Context, data *entity.PaymentCode) error
	GetAll(ctx context.Context, q utils.QueryParams) (utils.PaginationResult[entity.PaymentCode], error)
	GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.PaymentCode], error)
	FindByCode(ctx context.Context, code string, data *entity.PaymentCode) error
	FindById(ctx context.Context, id bson.ObjectID, data *entity.PaymentCode) error
	SetStatus(ctx context.Context, id bson.ObjectID, status constants.PaymentCodeStatus) error
	SetStatusByUserId(ctx context.Context, userId bson.ObjectID, code string, status constants.PaymentCodeStatus) error
}

type repository struct {
	collection *mongo.Collection
	refCode    *utils.ReferenceCode
	config     *config.Config
}

func NewRepository(db *mongo.Database, refCode *utils.ReferenceCode, config *config.Config) Repository {
	return &repository{
		collection: db.Collection("payment_code"),
		refCode:    refCode,
		config:     config,
	}
}

func (r *repository) Create(ctx context.Context, data *entity.PaymentCode) error {
	refCode, err := r.refCode.Next(constants.TypePayment.Short())
	if err != nil {
		return err
	}
	data.Code = refCode
	data.CreatedAt = time.Now()
	data.ExpiresAt = time.Now().Add(r.config.MaxPaymentTIme)

	res, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		data.ID = oid
	}

	return nil
}

func (r *repository) GetAll(ctx context.Context, q utils.QueryParams) (utils.PaginationResult[entity.PaymentCode], error) {
	filter := bson.M{}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		filter = bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"code": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"note": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
	}

	res, err := utils.Paginate[entity.PaymentCode](ctx, r.collection, utils.PaginationParam{
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

func (r *repository) GetAllByUserId(ctx context.Context, userId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.PaymentCode], error) {
	filter := bson.M{}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		filter = bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"code": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"note": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
		}
	} else {
		filter = bson.M{
			"user_id": userId,
		}
	}

	res, err := utils.Paginate[entity.PaymentCode](ctx, r.collection, utils.PaginationParam{
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

func (r *repository) FindByCode(ctx context.Context, code string, data *entity.PaymentCode) error {
	filter := bson.M{
		"code": code,
	}

	res := r.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}

	if err := res.Decode(data); err != nil {
		return err
	}
	return nil
}

func (r *repository) FindById(ctx context.Context, id bson.ObjectID, data *entity.PaymentCode) error {
	filter := bson.M{
		"_id": id,
	}

	res := r.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}

	if err := res.Decode(data); err != nil {
		return err
	}
	return nil
}

func (r *repository) SetStatus(ctx context.Context, id bson.ObjectID, status constants.PaymentCodeStatus) error {
	filter := bson.M{
		"_id": id,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) SetStatusByUserId(ctx context.Context, userId bson.ObjectID, code string, status constants.PaymentCodeStatus) error {
	filter := bson.M{
		"code":    code,
		"user_id": userId,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
