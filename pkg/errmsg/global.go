package errmsg

import "errors"

var (
	ErrInvalidJson        = errors.New("invalid json structure")
	ErrInternalServer     = errors.New("internal server error")
	ErrDataNotFound       = errors.New("data not found")
	ErrInfiniteLoop       = errors.New("error infinite loop")
	ErrEmailNotVerify     = errors.New("email not verified")
	ErrUnsupportProvider  = errors.New("unsupport the provider")
	ErrInvalidState       = errors.New("security token validation failed or session expired")
	ErrEmptyAuthCode      = errors.New("provider authorization code is empty")
	ErrAuthProcessFailed  = errors.New("authentication subsystem error")
	ErrMissingToken       = errors.New("missing token")
	ErrTokenAlreadyExists = errors.New("token is already exists")
)
