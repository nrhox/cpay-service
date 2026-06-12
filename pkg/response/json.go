package response

import (
	"net/http"

	"github.com/nrhox/cpay-service/pkg/rest"
	"github.com/nrhox/cpay-service/pkg/validation"
)

type ResJson struct {
	Data    any                     `json:"data,omitempty"`
	Message string                  `json:"message,omitempty"`
	Errors  []validation.ErrorField `json:"errors,omitempty"`
}

func Json(w http.ResponseWriter, httpCode int, r ResJson) {
	rest.JSON(w, httpCode, r)
}

type ResMetaPaginate struct {
	TotalPage int64 `json:"total_page"`
	TotalData int64 `json:"total_data"`
}

type ResJsonPaginate struct {
	ResJson
	Meta ResMetaPaginate `json:"meta"`
}

func JsonPaginate(w http.ResponseWriter, httpCode int, r ResJsonPaginate) {
	rest.JSON(w, httpCode, r)
}
