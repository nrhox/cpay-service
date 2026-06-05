package response

import (
	"net/http"

	"github.com/nrhox/cpay-service/pkg/validation"
)

func ValidationBody(w http.ResponseWriter, request any) bool {
	if errs := validation.New().Struct(request); len(errs) > 0 {
		Json(w, http.StatusBadRequest, ResJson{
			Errors: errs,
		})
		return false
	}
	return true
}
