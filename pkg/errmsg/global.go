package errmsg

import "errors"

var (
	ErrInvalidJson        = errors.New("invalid json structure")
	ErrInternalServer     = errors.New("internal server error")
	ErrDataNotFound       = errors.New("data not found")
	ErrInfiniteLoop       = errors.New("error infinite loop")
	ErrMissingToken       = errors.New("missing token")
	ErrTokenAlreadyExists = errors.New("token is already exists")
)
