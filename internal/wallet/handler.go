package wallet

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/rest"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	walletSvc *Service
	log       *slog.Logger
}

func NewHandler(
	walletSvc *Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		walletSvc: walletSvc,
		log:       log,
	}
}

func (h *Handler) NewWallet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	var req CreateWallet

	if err := rest.BindJson(r.Body, &req); err != nil {
		response.ParseError(w, errmsg.ErrInvalidJson, h.log)
		return
	}

	if ok := response.ValidationBody(w, req); !ok {
		return
	}

	wallet, err := h.walletSvc.CreateWallet(ctx, payload.UserID, req)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusCreated, response.ResJson{
		Data:    wallet,
		Message: "success create wallet",
	})
}

func (h *Handler) GetMyWallet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	wallets, err := h.walletSvc.GetAllWalletByUserID(ctx, payload.UserID)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    wallets,
		Message: "success get all wallets",
	})
}

func (h *Handler) GetWalletUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pUserId := chi.URLParam(r, "id")

	userId, err := bson.ObjectIDFromHex(pUserId)
	if err != nil {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	wallets, err := h.walletSvc.GetAllWalletByUserID(ctx, userId)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    wallets,
		Message: "success get all wallets",
	})
}

func (h *Handler) SetPrimaryWallet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	var req SetPrimaryWallet

	if err := rest.BindJson(r.Body, &req); err != nil {
		response.ParseError(w, errmsg.ErrInvalidJson, h.log)
		return
	}

	if ok := response.ValidationBody(w, req); !ok {
		return
	}

	if err := h.walletSvc.SetPrimary(ctx, payload.UserID, req); err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Message: "success set primary",
	})
}
