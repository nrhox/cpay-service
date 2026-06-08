package payment_code

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nrhox/cpay-service/internal/constants"
	"github.com/nrhox/cpay-service/internal/delivery/middleware"
	"github.com/nrhox/cpay-service/pkg/errmsg"
	"github.com/nrhox/cpay-service/pkg/response"
	"github.com/nrhox/cpay-service/pkg/rest"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	paymentService *Service
	log            *slog.Logger
}

func NewHandler(
	paymentService *Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		paymentService: paymentService,
		log:            log,
	}
}

func (h *Handler) CreatePaymentCode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	var req CreatePaymentCode

	if err := rest.BindJson(r.Body, &req); err != nil {
		response.ParseError(w, errmsg.ErrInvalidJson, h.log)
		return
	}

	if ok := response.ValidationBody(w, req); !ok {
		return
	}

	paymentCode, err := h.paymentService.CreatePaymentCode(ctx, payload.UserID, req)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusCreated, response.ResJson{
		Message: "Success create",
		Data:    paymentCode,
	})
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

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

	users, meta, err := h.paymentService.GetAll(ctx, utils.QueryParams{
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
			Data:    users,
			Message: "Success get all payment code",
		},
		Meta: meta,
	})
}

func (h *Handler) GetAllByUserId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pId := chi.URLParam(r, "id")

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	userId, err := bson.ObjectIDFromHex(pId)
	if err != nil {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	if payload.UserID == userId {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
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

	users, meta, err := h.paymentService.GetAllByUserId(ctx, userId, utils.QueryParams{
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
			Data:    users,
			Message: "Success get all payment code",
		},
		Meta: meta,
	})
}

func (h *Handler) GetAllMyCode(w http.ResponseWriter, r *http.Request) {
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

	users, meta, err := h.paymentService.GetAllByUserId(ctx, payload.UserID, utils.QueryParams{
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
			Data:    users,
			Message: "Success get all payment code",
		},
		Meta: meta,
	})
}

func (h *Handler) FindById(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pId := chi.URLParam(r, "id")

	paymentId, err := bson.ObjectIDFromHex(pId)
	if err != nil {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	payment, err := h.paymentService.FindById(ctx, paymentId)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    payment,
		Message: "success get payment code",
	})
}

func (h *Handler) FindByCode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pCode := chi.URLParam(r, "code")
	if len(pCode) != 14 || !strings.HasPrefix(pCode, constants.TypePayment.Short()) {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	payment, err := h.paymentService.FindByCode(ctx, pCode)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    payment,
		Message: "success get payment code",
	})
}

func (h *Handler) SetCancelByAdmin(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pId := chi.URLParam(r, "id")

	paymentId, err := bson.ObjectIDFromHex(pId)
	if err != nil {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	if err := h.paymentService.SetCancelByAdmin(ctx, paymentId); err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Message: "ok",
	})
}

func (h *Handler) SetCancelByUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	pCode := chi.URLParam(r, "code")
	if len(pCode) != 14 || !strings.HasPrefix(pCode, constants.TypePayment.Short()) {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	if err := h.paymentService.SetCancelByUser(ctx, payload.UserID, pCode); err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Message: "ok",
	})
}
