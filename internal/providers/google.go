package providers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config *oauth2.Config
}

type googleUserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewGoogleProvider(cfg *config.OauthConfig) {
	googleProv := &GoogleProvider{
		config: &oauth2.Config{
			RedirectURL:  cfg.RedirectURL,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}

	Register("google", googleProv)
}

func (g *GoogleProvider) GetLoginURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g *GoogleProvider) ExchangeCodeForUser(ctx context.Context, code string) (*Profile, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo?access_token="+token.AccessToken, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res googleUserResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if !res.VerifiedEmail {
		return nil, errmsg.ErrOauthEmailNotVerify
	}

	return &Profile{
		Email:        res.Email,
		FullName:     res.Name,
		Picture:      res.Picture,
		ProviderName: "google",
		ProviderID:   res.ID,
	}, nil
}
