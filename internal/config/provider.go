package config

type Providers struct {
	Google *OauthConfig
	Github *OauthConfig
}

type OauthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}
