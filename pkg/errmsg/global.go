package errmsg

import "errors"

var (
	ErrInvalidJson        = errors.New("invalid json structure")
	ErrInternalServer     = errors.New("internal server error")
	ErrDataNotFound       = errors.New("data not found")
	ErrInfiniteLoop       = errors.New("error infinite loop")
	ErrMissingToken       = errors.New("missing token")
	ErrTokenAlreadyExists = errors.New("token is already exists")

	ErrUserNotFound               = errors.New("user not found")
	ErrWalletNotFound             = errors.New("wallet not found")
	ErrBalanceDecreases           = errors.New("maaf saldo anda tidak cukup")
	ErrDestionationWalletNotFound = errors.New("maaf rekening tujuan tidak ditemukan")
	ErrPinNoMatch                 = errors.New("pin tidak cocok")
	ErrUncomplateForm             = errors.New("formulir tidak lengkap")
	ErrPaymentCodeNotFound        = errors.New("kode pembayaran tidak ditemukan")
	ErrMaxCreatedWallet           = errors.New("telah mencapai batas maksimal")
)
