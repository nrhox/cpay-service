package rest

import (
	"encoding/json"
	"io"
)

func BindJson(body io.Reader, req any) error {
	return json.NewDecoder(body).Decode(&req)
}
