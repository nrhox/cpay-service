package wallet

type CreateWallet struct {
	Name string `json:"wallet_name" validate:"required,alphanumspace"`
}
