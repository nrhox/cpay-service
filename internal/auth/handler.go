package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/securecookie"
	"github.com/nrhox/cpay-service/internal/config"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/internal/providers"
	"github.com/nrhox/cpay-service/internal/wallet"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/rest"
	"github.com/nrhox/cpay-service/pkg/security"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	authSvc       *Service
	log           *slog.Logger
	sessionConfig *config.Session
	frontendUrl   string
	tokenManager  *security.TokenManager
	sCookie       *securecookie.SecureCookie
}

func NewHandler(
	authSvc *Service,
	log *slog.Logger,
	sessionConfig *config.Session,
	frontendUrl string,
	tokenManager *security.TokenManager,
	sCookie *securecookie.SecureCookie,
) *Handler {
	return &Handler{
		authSvc:       authSvc,
		log:           log,
		sessionConfig: sessionConfig,
		frontendUrl:   frontendUrl,
		tokenManager:  tokenManager,
		sCookie:       sCookie,
	}
}

func (h *Handler) RedirectToFrontendError(w http.ResponseWriter, r *http.Request, errorKey string) {
	u, err := url.Parse(h.frontendUrl + constants.LOGIN_PAGE)
	if err != nil {
		response.ParseError(w, errmsg.ErrInternalServer, nil)
		return
	}

	q := u.Query()
	q.Set("error", errorKey)

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

	b := make([]byte, 16)
	rand.Read(b)
	stateString := base64.RawURLEncoding.EncodeToString(b)

	encode, err := h.sCookie.Encode(security.COOKIE_OAUTH_STATE, stateString)
	if err != nil {
		h.log.Error(err.Error())
		h.RedirectToFrontendError(w, r, "err_oauth_auth_process_failed")
		return
	}

	security.SetOauthState(w, h.sessionConfig.OauthStateDuration, encode)
	redirectUrl := provider.GetLoginURL(stateString)

	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

var oauthErrorMapper = map[string]map[string]string{
	"google": {
		"access_denied": "err_oauth_google_limit_or_cancel",

		"invalid_request": "err_oauth_google_bad_request",

		"invalid_scope": "err_oauth_google_invalid_scope",

		"unsupported_response_type": "err_oauth_google_unsupported_type",

		"server_error": "err_oauth_google_server_down",

		"temporarily_unavailable": "err_oauth_google_retry_later",

		"invalid_grant":         "err_oauth_google_code_expired_or_used",
		"invalid_client":        "err_oauth_google_wrong_client_credentials",
		"redirect_uri_mismatch": "err_oauth_google_callback_url_mismatch",
	},
	"github": {
		"access_denied":         "err_oauth_github_user_cancelled",
		"application_suspended": "err_oauth_github_suspended",
		"redirect_uri_mismatch": "err_oauth_github_callback_url_mismatch",
		"invalid_request":       "err_oauth_github_bad_request",
		"invalid_scope":         "err_oauth_github_invalid_scope",
	},
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var profile *providers.Profile

	userInject, _ := middleware.GetUserInjection(ctx)

	if userInject != nil {
		profile = userInject
	}

	if userInject == nil {
		providerName := chi.URLParam(r, "provider")

		provider := providers.Get(providerName)
		if provider == nil {
			h.RedirectToFrontendError(w, r, "err_oauth_unsupport")
			return
		}

		if oauthErr := r.FormValue("error"); oauthErr != "" {
			h.log.Error("OAuth provider %s returned error: %s", providerName, oauthErr)

			finalGenericKey := "err_oauth_provider_" + oauthErr

			if providerErrors, providerExists := oauthErrorMapper[providerName]; providerExists {
				if genericKey, errorExists := providerErrors[oauthErr]; errorExists {
					finalGenericKey = genericKey
				} else {
					finalGenericKey = "err_oauth_auth_process_failed"
				}
			} else {
				finalGenericKey = "err_oauth_auth_process_failed"
			}

			h.RedirectToFrontendError(w, r, finalGenericKey)
			return
		}

		state := r.FormValue("state")
		if state == "" {
			h.RedirectToFrontendError(w, r, "err_oauth_invalid_state")
			return
		}

		oauthStateCookie := security.GetOauthState(r)
		if oauthStateCookie == "" {
			h.RedirectToFrontendError(w, r, "err_oauth_invalid_state")
			return
		}

		var stateFromSession string
		if err := h.sCookie.Decode(security.COOKIE_OAUTH_STATE, oauthStateCookie, &stateFromSession); err != nil {
			h.RedirectToFrontendError(w, r, "err_oauth_invalid_state")
			return
		}

		if stateFromSession != state {
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
	}

	session, isComplate, err := h.authSvc.LoginUser(ctx, profile)
	if err != nil {
		if errors.Is(err, errmsg.ErrAccountSuspend) {
			h.RedirectToFrontendError(w, r, "err_account_suspend")
			return
		}
		h.RedirectToFrontendError(w, r, "err_oauth_auth_process_failed")
		return
	}

	security.DeleteOauthState(w)
	security.SetRefreshToken(w, h.sessionConfig.RefreshDuration, session.Token+"."+strings.ToUpper(session.ID.Hex()))

	if userInject == nil {
		if isComplate {
			http.Redirect(w, r, h.frontendUrl, http.StatusTemporaryRedirect)
			return
		}

		http.Redirect(w, r, h.frontendUrl+constants.INCOMPLATE_PAGE, http.StatusTemporaryRedirect)
		return
	}

	if isComplate {
		response.Json(w, http.StatusOK, response.ResJson{
			Data: "ok",
		})
		return
	}
	response.Json(w, http.StatusOK, response.ResJson{
		Data: "incomplate",
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	credential, err := middleware.GetAuthCredential(ctx)
	if err != nil {
		security.DeleteRefreshToken(w)
		security.DeleteAccessToken(w)
		response.ParseError(w, err, h.log)
		return
	}

	user, err := h.authSvc.RefreshToken(ctx, credential.Id, credential.Token)
	if err != nil {
		if errors.Is(err, errmsg.ErrUserNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			security.DeleteRefreshToken(w)
			security.DeleteAccessToken(w)
		}
		response.ParseError(w, err, h.log)
		return
	}

	if user.Status == constants.UserUncomplateRegister {
		response.Json(w, http.StatusBadRequest, response.ResJson{
			Message: errmsg.ErrInComplateUserRegister.Error(),
		})
		return
	}

	accessToken, err := h.tokenManager.Sign(security.AuthPayload{
		UserID: user.ID.Hex(),
		RoleId: user.RoleID,
	}, h.sessionConfig.AccessTokenDuration)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	security.SetAccessToken(w, h.sessionConfig.AccessTokenDuration, accessToken)

	response.Json(w, http.StatusOK, response.ResJson{
		Data: accessToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	credential, err := middleware.GetAuthCredential(ctx)
	if err != nil {
		response.Json(w, http.StatusOK, response.ResJson{
			Data: "ok",
		})
		return
	}

	if err := h.authSvc.Logout(ctx, credential.Id, credential.Token); err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	security.DeleteAccessToken(w)
	security.DeleteRefreshToken(w)

	response.Json(w, http.StatusOK, response.ResJson{
		Data: "ok",
	})
}

func (h *Handler) IncomplateRegister(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req wallet.CreateWallet

	if err := rest.BindJson(r.Body, &req); err != nil {
		response.ParseError(w, errmsg.ErrInvalidJson, h.log)
		return
	}

	if !response.ValidationBody(w, req) {
		return
	}

	credential, err := middleware.GetAuthCredential(ctx)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	if err := h.authSvc.IncomplateRegister(ctx, credential.Id, credential.Token, req); err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusCreated, response.ResJson{
		Message: "complate the form",
	})
}
