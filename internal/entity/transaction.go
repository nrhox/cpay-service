package entity

import (
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransactionParty struct {
	UserID        bson.ObjectID `bson:"user_id" json:"user_id"`
	WalletID      bson.ObjectID `bson:"wallet_id" json:"wallet_id"`
	WalletName    string        `bson:"wallet_name" json:"wallet_name"`
	AccountName   string        `bson:"account_name" json:"account_name"`
	AccountNumber string        `bson:"account_number" json:"account_number"`

	BalanceBefore uint64  `bson:"balance_before" json:"balance_before"`
	BalanceAfter  *uint64 `bson:"balance_after,omitempty" json:"balance_after,omitempty"`
}

type Transaction struct {
	ID        bson.ObjectID             `bson:"_id,omitempty" json:"id"`
	Reference string                    `bson:"reference" json:"reference"`
	Type      constants.TransactionType `bson:"type" json:"type"`
	Title     string                    `bson:"title" json:"title"`

	Amount   uint64                      `bson:"amount" json:"amount"`
	Currency string                      `bson:"currency" json:"currency"`
	Status   constants.TransactionStatus `bson:"status" json:"status"`

	Source      *TransactionParty `bson:"source,omitempty" json:"source,omitempty"`
	Destination *TransactionParty `bson:"destination,omitempty" json:"destination,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
