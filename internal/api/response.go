package api

import (
	"encoding/json"
	"net/http"

	"github.com/retail-core/auth-service/internal/common"
	"github.com/retail-core/auth-service/internal/logger"
	"go.uber.org/zap"
)

func WriteJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, err error) {
	appError, ok := common.IsAppError(err)
	if !ok {
		logger.L().Error("Internal Error", zap.Error(err))
		appError = common.ErrInternal
	}
	WriteJson(w, appError.Status, appError)
}