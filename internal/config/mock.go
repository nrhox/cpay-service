package config

import "github.com/nrhox/cpay-service/internal/entity"

type UserMock struct {
	User     *entity.User
	WithMock bool
	MockFile string
}
