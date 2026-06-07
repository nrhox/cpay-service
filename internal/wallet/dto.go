package wallet

import "go.mongodb.org/mongo-driver/v2/bson"

type CreateWallet struct {
	Name string `json:"wallet_name" validate:"required,alphanumspace"`
}

type SetPrimaryWallet struct {
	WalletId bson.ObjectID `json:"wallet_id" validate:"required"`
}
