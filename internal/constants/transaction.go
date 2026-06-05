package constants

type (
	TransactionType   string
	TransactionStatus string
)

const (
	TypeTopup    TransactionType = "TOPUP"
	TypeTransfer TransactionType = "TRANSFER"
	TypePayment  TransactionType = "PAYMENT"

	StatusPending   TransactionStatus = "PENDING"
	StatusSuccess   TransactionStatus = "SUCCESS"
	StatusFailed    TransactionStatus = "FAILED"
	StatusCancelled TransactionStatus = "CANCELLED"
)
