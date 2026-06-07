package topup_request

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/rest"
)

type Handler struct {
	log      *slog.Logger
	topupSvc *Service
}

func NewHandler(
	topupSvc *Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		log:      log,
		topupSvc: topupSvc,
	}
}

func (h *Handler) RequestTopup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	var req RequestTopup

	if err := rest.BindJson(r.Body, &req); err != nil {
		response.ParseError(w, errmsg.ErrInvalidJson, h.log)
		return
	}

	if ok := response.ValidationBody(w, req); !ok {
		return
	}

	topUp, err := h.topupSvc.CreateRequest(ctx, user.UserID, req)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusCreated, response.ResJson{
		Data:    topUp,
		Message: "Success create request",
	})
}
