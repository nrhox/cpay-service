package transaction

import (
	"context"
	"time"

	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	Create(ctx context.Context, entity *entity.Transaction) error
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
