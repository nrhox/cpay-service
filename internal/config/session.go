package config

import "time"

type Session struct {
	RefreshDuration     time.Duration
	AccessTokenDuration time.Duration
	JwtPrivateKey       string
	JwtPublicKey        string
	SaltKey             string
}
