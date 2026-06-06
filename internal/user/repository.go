package user

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

type Repository interface {
	NewUser(ctx context.Context, entity *entity.User) error
	GetOneByEmail(ctx context.Context, email string, entity *entity.User) error
	UpsertProvider(ctx context.Context, id bson.ObjectID, prov entity.AuthProvider) error
	GetOneNoSuspendById(ctx context.Context, id bson.ObjectID, entity *entity.User) error
	CheckUserStatus(ctx context.Context, id bson.ObjectID, status constants.UserStatus) (bool, error)
	SetStatus(ctx context.Context, id bson.ObjectID, status constants.UserStatus) error
	GetAll(ctx context.Context, notId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.User], error)
	GetOneById(ctx context.Context, id bson.ObjectID, entity *entity.User) error
}

var fieldAllowSort []string = []string{
	"_id",
	"full_name",
	"role",
	"email",
	"created_at",
	"updated_at",
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

func (r *repository) GetOneById(ctx context.Context, id bson.ObjectID, entity *entity.User) error {
	filter := bson.M{
		"_id": id,
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

func (r *repository) CheckUserStatus(ctx context.Context, id bson.ObjectID, status constants.UserStatus) (bool, error) {
	filter := bson.M{
		"_id":    id,
		"status": status,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (r *repository) SetStatus(ctx context.Context, id bson.ObjectID, status constants.UserStatus) error {
	filter := bson.M{
		"_id": id,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	if _, err := r.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (r *repository) GetAll(ctx context.Context, notId bson.ObjectID, q utils.QueryParams) (utils.PaginationResult[entity.User], error) {
	filter := bson.M{}
	excludeFilter := bson.M{"$ne": notId}

	if q.SearchKeyword != "" {
		likeStartKeyword := "^" + q.SearchKeyword
		filter = bson.M{
			"$or": bson.A{
				bson.M{"_id": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"full_name": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
				bson.M{"email": bson.M{"$regex": likeStartKeyword, "$options": "i"}},
			},
			"_id": excludeFilter,
		}
	} else {
		filter = bson.M{
			"_id": excludeFilter,
		}
	}

	res, err := utils.Paginate[entity.User](ctx, r.collection, utils.PaginationParam{
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
