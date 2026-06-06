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

func (t TransactionType) Short() string {
	switch t {
	case TypeTopup:
		return "TP"
	case TypeTransfer:
		return "TF"
	case TypePayment:
		return "PY"
	default:
		return "TX"
	}
}
