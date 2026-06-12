package topup_request

type RequestTopup struct {
	WalletNumber string `json:"wallet_number" validate:"required,number,len=12"`
	Amount       uint64 `json:"amount" validate:"required,number"`
}
