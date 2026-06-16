package config

import "time"

type Session struct {
	RefreshDuration     time.Duration
	AccessTokenDuration time.Duration
	OauthStateDuration  time.Duration
	JwtPrivateKey       string
	JwtPublicKey        string
	HashKey             string
	BlocKey             string
}
