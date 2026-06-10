package transaction

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/utils"
)

type Handler struct {
	tranactionSvc *Service
	log           *slog.Logger
}

func NewHandler(
	tranactionSvc *Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		tranactionSvc: tranactionSvc,
		log:           log,
	}
}

func (h *Handler) GetMyTransaction(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	queryParams := r.URL.Query()

	qKeyword := queryParams.Get("q")
	qSort := queryParams.Get("sort")
	qOrder := queryParams.Get("order_by")

	pageStr := queryParams.Get("page")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page <= 0 {
		page = 1
	}

	limitStr := queryParams.Get("limit")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 10
	}

	transaction, meta, err := h.tranactionSvc.GetAllByUserId(ctx, payload.UserID, utils.QueryParams{
		Page:          page,
		Limit:         limit,
		SortBy:        qSort,
		SortOrder:     qOrder,
		SearchKeyword: qKeyword,
	})
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.JsonPaginate(w, http.StatusOK, response.ResJsonPaginate{
		ResJson: response.ResJson{
			Data:    transaction,
			Message: "Success get all",
		},
		Meta: meta,
	})
}

func (h *Handler) GetMyTransactionByAccountNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	accountNumber := chi.URLParam(r, "account_number")

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	queryParams := r.URL.Query()

	qKeyword := queryParams.Get("q")
	qSort := queryParams.Get("sort")
	qOrder := queryParams.Get("order_by")

	pageStr := queryParams.Get("page")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page <= 0 {
		page = 1
	}

	limitStr := queryParams.Get("limit")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 10
	}

	transactions, meta, err := h.tranactionSvc.GetAllByAccountNumber(ctx, accountNumber, payload.UserID, utils.QueryParams{
		Page:          page,
		Limit:         limit,
		SortBy:        qSort,
		SortOrder:     qOrder,
		SearchKeyword: qKeyword,
	})
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.JsonPaginate(w, http.StatusOK, response.ResJsonPaginate{
		ResJson: response.ResJson{
			Data:    transactions,
			Message: "Success get all",
		},
		Meta: meta,
	})
}

func (h *Handler) GetOneByRefCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	refCode := chi.URLParam(r, "ref_code")
	if len(refCode) != 14 {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	transaction, err := h.tranactionSvc.GetOneByRefAndUserId(ctx, refCode, payload.UserID)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    transaction,
		Message: "Success get one",
	})
}
