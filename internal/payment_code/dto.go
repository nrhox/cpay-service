package payment_code

import "go.mongodb.org/mongo-driver/v2/bson"

type CreatePaymentCode struct {
	WalletId bson.ObjectID `json:"wallet_id" validate:"required"`
	Amount   uint64        `json:"amount" validate:"required,number"`
	Note     string        `json:"note" validate:"omitempty,max=50"`
}

type CreetePayingTransaction struct {
	WalletId    bson.ObjectID `json:"wallet_id" validate:"required"`
	PaymentCode string        `json:"payment_code" validate:"required,len=14,alphanum"`
	Pin         string        `json:"pin" validate:"required,len=6,number"`
}
