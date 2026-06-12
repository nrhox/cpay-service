package constants

type PaymentCodeStatus string

const (
	PaymentActive    PaymentCodeStatus = "ACTIVE"
	PaymentPaid      PaymentCodeStatus = "PAID"
	PaymentCancelled PaymentCodeStatus = "CANCELLED"
	PaymentExpired   PaymentCodeStatus = "EXPIRED"
)
