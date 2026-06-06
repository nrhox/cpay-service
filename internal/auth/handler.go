package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/providers"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/security"
)

type Handler struct {
	authSvc       *Service
	log           *slog.Logger
	sessionConfig *config.Session
	frontendUrl   string
	tokenManager  *security.TokenManager
}

func NewHandler(
	authSvc *Service,
	log *slog.Logger,
	sessionConfig *config.Session,
	frontendUrl string,
	tokenManager *security.TokenManager,
) *Handler {
	return &Handler{
		authSvc:       authSvc,
		log:           log,
		sessionConfig: sessionConfig,
		frontendUrl:   frontendUrl,
		tokenManager:  tokenManager,
	}
}

func (h *Handler) RedirectToFrontendError(w http.ResponseWriter, r *http.Request, reason string) {
	u, err := url.Parse(h.frontendUrl + constants.LOGIN_PAGE)
	if err != nil {
		response.ParseError(w, errmsg.ErrInternalServer, nil)
		return
	}

	q := u.Query()
	q.Set("reason", reason)

	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	providerName := chi.URLParam(r, "provider")

	provider := providers.Get(providerName)
	if provider == nil {
		h.RedirectToFrontendError(w, r, "err_oauth_unsupport")
		return
	}

	stateString := security.GenerateDynamicToken(h.sessionConfig.SaltKey) + "__" + providerName
	redirectUrl := provider.GetLoginURL(stateString)

	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var profile *providers.Profile

	providerName := chi.URLParam(r, "provider")

	provider := providers.Get(providerName)
	if provider == nil {
		h.RedirectToFrontendError(w, r, "err_oauth_unsupport")
		return
	}

	state := r.FormValue("state")
	state = strings.TrimSuffix(state, "__"+providerName)
	if !security.ValidateDynamicToken(state, h.sessionConfig.SaltKey) {
		h.RedirectToFrontendError(w, r, "err_oauth_invalid_state")
		return
	}

	code := r.FormValue("code")
	if code == "" {
		h.RedirectToFrontendError(w, r, "err_oauthh_empty_auth_code")
		return
	}

	profileData, err := provider.ExchangeCodeForUser(ctx, code)
	if err != nil {
		h.log.Error(err.Error())
		h.RedirectToFrontendError(w, r, "err_oauth_auth_process_failed")
		return
	}

	profile = profileData

	session, isComplate, err := h.authSvc.LoginUser(ctx, profile)
	if err != nil {
		h.RedirectToFrontendError(w, r, "err_oauth_auth_process_failed")
		return
	}

	security.SetRefreshToken(w, h.sessionConfig.RefreshDuration, session.Token+"."+strings.ToUpper(session.ID.Hex()))

	if isComplate {
		http.Redirect(w, r, h.frontendUrl, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, h.frontendUrl+constants.INCOMPLATE_PAGE, http.StatusTemporaryRedirect)
	}
}
