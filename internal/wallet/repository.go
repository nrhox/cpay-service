package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	Create(ctx context.Context, wallet *entity.Wallet) error
	AvailableWalletPrimary(ctx context.Context, userId bson.ObjectID) (bool, error)
}

type repository struct {
	collection *mongo.Collection
	snowId     *utils.Snowflake
}

func NewRepository(db *mongo.Database, snowId *utils.Snowflake) Repository {
	return &repository{
		collection: db.Collection("wallets"),
		snowId:     snowId,
	}
}

func (r *repository) Create(ctx context.Context, wallet *entity.Wallet) error {
	wallet.Balance = 0
	wallet.AccountNumber = r.snowId.NextID()
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	res, err := r.collection.InsertOne(ctx, wallet)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		wallet.ID = oid
	}

	return nil
}

func (r *repository) AvailableWalletPrimary(ctx context.Context, userId bson.ObjectID) (bool, error) {
	filter := bson.M{
		"is_primary": true,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return true, nil
		}
		return false, err
	}

	if count == 0 {
		return true, nil
	}
	return false, nil
}
