package main

import (
	"log"
	"net/http"

	"github.com/retail-core/auth-service/internal/api"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	logger.Init(cfg.MODE)

	router := api.NewRouter()

	logger := logger.L()

	logger.Info("Auth service starting on Port", zap.String("port", cfg.PORT))
	err := http.ListenAndServe("0.0.0.0:"+cfg.PORT, router)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
