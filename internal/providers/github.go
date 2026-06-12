package providers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubProvider struct {
	config *oauth2.Config
}

type githubUserResponse struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

type githubEmailResponse struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewGitHubProvider(cfg *config.OauthConfig) {
	githubProj := &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: github.Endpoint,
		},
	}

	Register("github", githubProj)
}

func (h *GitHubProvider) GetLoginURL(state string) string {
	return h.config.AuthCodeURL(state)
}

func (h *GitHubProvider) ExchangeCodeForUser(ctx context.Context, code string) (*Profile, error) {
	token, err := h.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res githubUserResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	userEmail := res.Email
	if userEmail == "" {
		emailReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			return nil, err
		}
		emailReq.Header.Set("Authorization", "Bearer "+token.AccessToken)

		emailResp, err := http.DefaultClient.Do(emailReq)
		if err != nil {
			return nil, err
		}
		defer emailResp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errmsg.ErrGithubApi
		}

		emailBody, err := io.ReadAll(emailResp.Body)
		if err != nil {
			return nil, err
		}

		var emails []githubEmailResponse
		if err := json.Unmarshal(emailBody, &emails); err != nil {
			return nil, err
		}

		for _, e := range emails {
			if e.Primary && e.Verified {
				userEmail = e.Email
				break
			}
		}
	}

	if userEmail == "" {
		return nil, errmsg.ErrOauthEmailNotVerify
	}

	fullName := res.Name
	if fullName == "" {
		fullName = res.Login
	}

	return &Profile{
		Email:        userEmail,
		FullName:     fullName,
		Picture:      res.AvatarURL,
		ProviderName: "github",
		ProviderID:   strconv.Itoa(res.ID),
	}, nil
}
