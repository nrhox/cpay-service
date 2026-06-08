package user

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
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	userSvc *Service
	log     *slog.Logger
}

func NewHandler(
	userSvc *Service,
	log *slog.Logger,
) *Handler {
	return &Handler{
		userSvc: userSvc,
		log:     log,
	}
}

func (h *Handler) GetAllUser(w http.ResponseWriter, r *http.Request) {
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

	users, meta, err := h.userSvc.GetAll(ctx, payload.UserID, utils.QueryParams{
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
			Message: "Success get all user",
		},
		Meta: meta,
	})
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userSvc.GetOne(ctx, userId)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    user,
		Message: "Success get user",
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payload, err := middleware.GetPayloadUser(ctx)
	if err != nil {
		response.ParseError(w, errmsg.ErrMissingToken, h.log)
		return
	}

	user, err := h.userSvc.GetOne(ctx, payload.UserID)
	if err != nil {
		response.ParseError(w, err, h.log)
		return
	}

	response.Json(w, http.StatusOK, response.ResJson{
		Data:    user,
		Message: "Success",
	})
}
