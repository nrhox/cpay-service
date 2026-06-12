package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/validation"
)

type ErrorWithResponse struct {
	Err        error
	Message    string
	StatusCode int
	Errors     []validation.ErrorField
}

var listError = []ErrorWithResponse{
	{
		Err:        errmsg.ErrInvalidJson,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrInvalidJson.Error(),
	},
	{
		Err:        errmsg.ErrDataNotFound,
		StatusCode: http.StatusNotFound,
		Message:    errmsg.ErrDataNotFound.Error(),
	},
	{
		Err:        errmsg.ErrOauthEmailNotVerify,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrOauthEmailNotVerify.Error(),
	},
	{
		Err:        errmsg.ErrOauthUnsupport,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrOauthUnsupport.Error(),
	},
	{
		Err:        errmsg.ErrMissingToken,
		StatusCode: http.StatusUnauthorized,
		Message:    errmsg.ErrMissingToken.Error(),
	},
	{
		Err:        jwt.ErrTokenMalformed,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrTokenMalformed.Error(),
	},
	{
		Err:        jwt.ErrTokenExpired,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrTokenExpired.Error(),
	},
	{
		Err:        jwt.ErrTokenNotValidYet,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrTokenNotValidYet.Error(),
	},
	{
		Err:        jwt.ErrSignatureInvalid,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrSignatureInvalid.Error(),
	},
	{
		Err:        jwt.ErrTokenInvalidId,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrTokenInvalidId.Error(),
	},
	{
		Err:        jwt.ErrTokenUsedBeforeIssued,
		StatusCode: http.StatusUnauthorized,
		Message:    jwt.ErrTokenUsedBeforeIssued.Error(),
	},
	{
		Err:        errmsg.ErrUserNotFound,
		StatusCode: http.StatusNotFound,
		Message:    errmsg.ErrUserNotFound.Error(),
	},
	{
		Err:        errmsg.ErrWalletNotFound,
		StatusCode: http.StatusNotFound,
		Message:    errmsg.ErrWalletNotFound.Error(),
	},
	{
		Err:        errmsg.ErrBalanceDecreases,
		StatusCode: http.StatusNotFound,
		Message:    errmsg.ErrBalanceDecreases.Error(),
	},
	{
		Err:        errmsg.ErrDestionationWalletNotFound,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrUncomplateForm.Error(),
		Errors: []validation.ErrorField{
			{
				Field:   "destination",
				Message: errmsg.ErrDestionationWalletNotFound.Error(),
			},
		},
	},
	{
		Err:        errmsg.ErrPinNoMatch,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrUncomplateForm.Error(),
		Errors: []validation.ErrorField{
			{
				Field:   "pin",
				Message: errmsg.ErrPinNoMatch.Error(),
			},
		},
	},
	{
		Err:        errmsg.ErrPaymentCodeNotFound,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrUncomplateForm.Error(),
		Errors: []validation.ErrorField{
			{
				Field:   "payment_code",
				Message: errmsg.ErrPaymentCodeNotFound.Error(),
			},
		},
	},
	{
		Err:        errmsg.ErrMaxCreatedWallet,
		StatusCode: http.StatusBadRequest,
		Message:    errmsg.ErrUncomplateForm.Error(),
		Errors: []validation.ErrorField{
			{
				Field:   "wallet_name",
				Message: errmsg.ErrMaxCreatedWallet.Error(),
			},
		},
	},
}

func ParseError(w http.ResponseWriter, err error, log *slog.Logger) {
	for _, target := range listError {
		if errors.Is(err, target.Err) {
			Json(w, target.StatusCode, ResJson{
				Message: target.Message,
				Errors:  target.Errors,
			})
			return
		}
	}
	if log != nil {
		log.Error(err.Error())
	}

	Json(w, http.StatusInternalServerError, ResJson{
		Message: errmsg.ErrInternalServer.Error(),
	})
}
