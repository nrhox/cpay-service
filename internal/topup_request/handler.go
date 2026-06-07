package topup_request

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
	"github.com/nrhox/cpay-service/pkg/rest"
	"github.com/nrhox/cpay-service/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (h *Handler) GetAllTopUp(w http.ResponseWriter, r *http.Request) {
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

	requests, meta, err := h.topupSvc.GetAll(ctx, utils.QueryParams{
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
			Data:    requests,
			Message: "Success get all requests",
		},
		Meta: meta,
	})
}

func (h *Handler) GetOneById(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	pId := chi.URLParam(r, "id")

	topupId, err := bson.ObjectIDFromHex(pId)
	if err != nil {
		response.ParseError(w, errmsg.ErrDataNotFound, h.log)
		return
	}

	topUp, err := h.topupSvc.GetOneById(ctx, topupId)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    topUp,
		Message: "Success get top up",
	})
}
