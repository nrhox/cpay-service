package entity

import (
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Wallet struct {
	ID            bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	UserID        bson.ObjectID          `bson:"user_id" json:"user_id"`
	Name          string                 `bson:"name" json:"name"`
	AccountNumber string                 `bson:"account_number" json:"account_number"`
	Balance       uint64                 `bson:"balance" json:"balance"`
	Status        constants.WalletStatus `bson:"status" json:"status"`
	IsPrimary     bool                   `bson:"is_primary" json:"is_primary"`
	Pin           string                 `bson:"pin" json:"pin,omitempty"`
	CreatedAt     time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at" json:"updated_at"`
}
