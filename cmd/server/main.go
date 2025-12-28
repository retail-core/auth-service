package main

import (
	"context"
	"log"
	"net/http"

	"github.com/retail-core/auth-service/internal/api"
	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/consumer"
	"github.com/retail-core/auth-service/internal/db"
	"github.com/retail-core/auth-service/internal/logger"
	"github.com/retail-core/auth-service/internal/queue"
	"github.com/retail-core/auth-service/internal/user"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	logger.Init(cfg.MODE)

	db, query := db.Connect(cfg.DB_SOURCE)
	userRepo := user.NewPGRepository(query, db)

	publisher := queue.NewPublisher(cfg.RABBITMQ_URL)
	_ = publisher.SetupDomainExchange()

	router := api.NewRouter(userRepo, publisher)

	logger := logger.L()

	go func() {
		logger.Info("Auth service starting on Port", zap.String("port", cfg.PORT))
		err := http.ListenAndServe("0.0.0.0:"+cfg.PORT, router)
		if err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	repo := user.NewPGRepository(query, db)
	authService := auth.NewService(cfg.JWT_SECRET_KEY, repo, *publisher)
	authConsumer := consumer.NewAuthConsumer(publisher.Ch, authService)
	ctx := context.Background()
	authConsumer.StartConsumption(ctx)

	select {}
}
