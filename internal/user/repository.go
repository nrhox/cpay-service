package user

import (
	"context"
	"errors"
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/entity"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	NewUser(ctx context.Context, entity *entity.User) error
	GetOneByEmail(ctx context.Context, email string, entity *entity.User) error
	UpsertProvider(ctx context.Context, id bson.ObjectID, prov entity.AuthProvider) error
	GetOneNoSuspendById(ctx context.Context, id bson.ObjectID, entity *entity.User) error
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("users"),
	}
}

func (r *repository) NewUser(ctx context.Context, entity *entity.User) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	entity.Status = constants.UserUncomplateRegister

	res, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		entity.ID = oid
	}

	return nil
}

func (r *repository) GetOneByEmail(ctx context.Context, email string, entity *entity.User) error {
	filter := bson.M{
		"email": email,
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

func (r *repository) UpsertProvider(ctx context.Context, id bson.ObjectID, prov entity.AuthProvider) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$addToSet": bson.M{
			"oauth_providers": bson.M{
				"id":       prov.ID,
				"provider": prov.Provider,
			},
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errmsg.ErrDataNotFound
		}
		return err
	}

	return nil
}

func (r *repository) GetOneNoSuspendById(ctx context.Context, id bson.ObjectID, entity *entity.User) error {
	filter := bson.M{
		"_id": id,
		"status": bson.M{
			"$ne": constants.UserSuspended,
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
