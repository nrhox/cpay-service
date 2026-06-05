package entity

import (
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PaymentCode struct {
	ID        bson.ObjectID               `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID               `bson:"user_id" json:"user_id"`
	WalletID  bson.ObjectID               `bson:"wallet_id" json:"wallet_id"`
	Merchant  string                      `bson:"merchant" json:"merchant"`
	Code      string                      `bson:"code" json:"code"`
	Amount    uint64                      `bson:"amount" json:"amount"`
	Note      string                      `bson:"note" json:"note"`
	Status    constants.PaymentCodeStatus `bson:"status" json:"status"`
	ExpiresAt time.Time                   `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time                   `bson:"created_at" json:"created_at"`
}
