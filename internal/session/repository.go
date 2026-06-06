package session

import (
	"context"
	"errors"
	"time"

	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	Create(ctx context.Context, entity *entity.Session) error
	IsTokenAlready(ctx context.Context, token string) error
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("sessions"),
	}
}

func (r *repository) Create(ctx context.Context, entity *entity.Session) error {
	entity.CreatedAt = time.Now()

	res, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		entity.ID = oid
	}

	return nil
}

func (r *repository) IsTokenAlready(ctx context.Context, token string) error {
	filter := bson.M{
		"token": token,
	}
	res, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	if res > 0 {
		return errmsg.ErrTokenAlreadyExists
	}

	return nil
}
