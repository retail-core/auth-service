package api

import (
	"encoding/json"
	"net/http"
	"github.com/retail-core/auth-service/internal/common"
)

func WriteJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, err error) {
	appError, ok := common.IsAppError(err)
	if !ok {
		appError = common.ErrInternal
	}
	WriteJson(w, appError.Status, appError)
}