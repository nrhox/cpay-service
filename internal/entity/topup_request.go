package entity

import (
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TopupRequest struct {
	ID          bson.ObjectID               `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID               `bson:"user_id" json:"user_id"`
	WalletID    bson.ObjectID               `bson:"wallet_id" json:"wallet_id"`
	Amount      uint64                      `bson:"amount" json:"amount"`
	Reference   string                      `bson:"reference" json:"reference"`
	Status      constants.TransactionStatus `bson:"status" json:"status"`
	RequestedAt time.Time                   `bson:"requested_at" json:"requested_at"`
	ReviewedAt  *time.Time                  `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	AdminID     *bson.ObjectID              `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
}
