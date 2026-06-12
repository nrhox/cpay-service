package wallet

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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, wallet *entity.Wallet) error
	AvailableWalletPrimary(ctx context.Context, userId bson.ObjectID) (bool, error)
	FindActiveByAccounNumberWithUser(ctx context.Context, userId bson.ObjectID, accountNumber string, data *entity.Wallet) error
	FindByAccounNumberWithUser(ctx context.Context, accountNumber string, data *entity.WalletWithUser) error
	FindActiveByAccountNumber(ctx context.Context, accountNumber string, data *entity.Wallet) error
	FindByAccountNumber(ctx context.Context, accountNumber string, status constants.WalletStatus, data *entity.Wallet) error
	UpdateBalance(ctx context.Context, id bson.ObjectID, amount int) error
	FindById(ctx context.Context, id bson.ObjectID, data *entity.Wallet) error
	FindActiveByIdWithUser(ctx context.Context, id bson.ObjectID, userId bson.ObjectID, data *entity.Wallet) error
	SetAllStatusByUserId(ctx context.Context, userId bson.ObjectID, status constants.WalletStatus) error
	GetAllByUserId(ctx context.Context, userId bson.ObjectID) ([]entity.Wallet, error)
	SetOnePrimary(ctx context.Context, userId bson.ObjectID, walletId bson.ObjectID) error
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

const MAX_WALLET = 4

func (r *repository) Create(ctx context.Context, wallet *entity.Wallet) error {
	filter := bson.M{
		"user_id": wallet.UserID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if count >= MAX_WALLET {
		return errmsg.ErrMaxCreatedWallet
	}

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

func (r *repository) FindActiveByAccounNumberWithUser(ctx context.Context, userId bson.ObjectID, accountNumber string, data *entity.Wallet) error {
	filter := bson.M{
		"account_number": accountNumber,
		"status":         constants.WalletActive,
		"user_id":        userId,
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

func (r *repository) FindByAccounNumberWithUser(ctx context.Context, accountNumber string, data *entity.WalletWithUser) error {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"account_number": accountNumber,
			"status":         constants.WalletActive,
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user_array",
		}}},

		{{Key: "$unwind", Value: bson.M{
			"path":                       "$user_array",
			"preserveNullAndEmptyArrays": false,
		}}},
		{{Key: "$match", Value: bson.M{
			"user_array.status": constants.UserActive,
		}}},

		{{Key: "$addFields", Value: bson.M{
			"user": "$user_array",
		}}},

		{{Key: "$project", Value: bson.M{
			"pin":                  0,
			"user_array":           0,
			"user.oauth_providers": 0,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		if err := cursor.Decode(data); err != nil {
			return err
		}
		return nil
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return mongo.ErrNoDocuments
}

func (r *repository) FindActiveByAccountNumber(ctx context.Context, accountNumber string, data *entity.Wallet) error {
	filter := bson.M{
		"account_number": accountNumber,
		"status":         constants.WalletActive,
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

func (r *repository) FindByAccountNumber(ctx context.Context, accountNumber string, status constants.WalletStatus, data *entity.Wallet) error {
	filter := bson.M{
		"account_number": accountNumber,
		"status":         status,
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

func (r *repository) FindById(ctx context.Context, id bson.ObjectID, data *entity.Wallet) error {
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

func (r *repository) UpdateBalance(ctx context.Context, id bson.ObjectID, amount int) error {
	filter := bson.M{
		"_id": id,
	}

	if amount < 0 {
		filter["balance"] = bson.M{"$gte": -amount}
	}

	update := bson.M{
		"$inc": bson.M{
			"balance": amount,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 && amount < 0 {
		return errmsg.ErrBalanceDecreases
	}

	return nil
}

func (r *repository) SetAllStatusByUserId(ctx context.Context, userId bson.ObjectID, status constants.WalletStatus) error {
	filter := bson.M{
		"user_id": userId,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAllByUserId(ctx context.Context, userId bson.ObjectID) ([]entity.Wallet, error) {
	filter := bson.M{"user_id": userId}
	projection := bson.M{"pin": 0}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []entity.Wallet
	if err := cursor.All(ctx, &wallets); err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *repository) SetOnePrimary(ctx context.Context, userId bson.ObjectID, walletId bson.ObjectID) error {
	filterWallet := bson.M{"_id": walletId, "user_id": userId}
	updateWallet := bson.M{"$set": bson.M{"is_primary": true}}

	resSpesific, err := r.collection.UpdateOne(ctx, filterWallet, updateWallet)
	if err != nil {
		return err
	}

	if resSpesific.MatchedCount == 0 {
		return errmsg.ErrWalletNotFound
	}

	if resSpesific.ModifiedCount == 0 {
		return nil
	}

	filterAllWallet := bson.M{
		"user_id": userId,
		"_id": bson.M{
			"$ne": walletId,
		},
	}
	update := bson.M{"$set": bson.M{"is_primary": false}}

	_, err = r.collection.UpdateMany(ctx, filterAllWallet, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) FindActiveByIdWithUser(ctx context.Context, walletId bson.ObjectID, userId bson.ObjectID, data *entity.Wallet) error {
	filter := bson.M{"_id": walletId, "user_id": userId, "status": constants.WalletActive}
	res := r.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}

	if err := res.Decode(data); err != nil {
		return err
	}
	return nil
}
