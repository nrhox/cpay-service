package constants

type (
	Role       uint8
	UserStatus string
)

const (
	RoleUser  Role = 1 << 0
	RoleAdmin Role = 1 << 1
)

const (
	UserActive             UserStatus = "ACTIVE"
	UserSuspended          UserStatus = "SUSPENDED"
	UserUncomplateRegister UserStatus = "UNCOMPLATE"
)
