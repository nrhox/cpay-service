package wallet

import "go.mongodb.org/mongo-driver/v2/bson"

type CreateWallet struct {
	Name string `json:"wallet_name" validate:"required,alphanumspace"`
}

type SetPrimaryWallet struct {
	WalletId bson.ObjectID `json:"wallet_id" validate:"required"`
}

type TransferBalance struct {
	WalletId           bson.ObjectID `json:"wallet_id" validate:"required"`
	DestionationWallet string        `json:"destination" validate:"required,number,len=12"`
	Amount             uint64        `json:"amount" validate:"required,number"`
	Note               string        `json:"note" validate:"omitempty,max=50"`
}
