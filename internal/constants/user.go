package constants

type (
	Role       string
	UserStatus string
)

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"

	UserActive    UserStatus = "ACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
)
