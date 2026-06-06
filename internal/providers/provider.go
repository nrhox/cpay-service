package providers

import (
	"context"
)

type Profile struct {
	Email        string
	FullName     string
	Picture      string
	ProviderName string
	ProviderID   string
}

type AuthProvider interface {
	GetLoginURL(state string) string
	ExchangeCodeForUser(ctx context.Context, code string) (*Profile, error)
}

var activeProviders = make(map[string]AuthProvider)

func Register(name string, p AuthProvider) {
	activeProviders[name] = p
}

func Get(name string) AuthProvider {
	return activeProviders[name]
}
