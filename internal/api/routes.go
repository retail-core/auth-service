package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/db"
	"github.com/retail-core/auth-service/internal/middleware"
	"github.com/retail-core/auth-service/internal/queue"
	"github.com/retail-core/auth-service/internal/user"
)

func NewRouter() http.Handler {
	cfg := config.LoadConfig() // remove this guy later already in main
	router := chi.NewRouter()

	db, query := db.Connect(cfg.DB_SOURCE)
	userRepo := user.NewPGRepository(query, db)

	publisher := queue.NewPublisher(cfg.RABBITMQ_URL)
	_ = publisher.SetupDomainExchange()

	authService := auth.NewService(cfg.JWT_SECRET_KEY, userRepo, *publisher)
	authHandler := NewAuthHandler(authService)

	router.Use(middleware.RequestLogger)

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	router.Route("/auth", func(v1 chi.Router) {
		v1.Route("/v1", func(auth chi.Router) {
			auth.Post("/login", authHandler.Login)
			auth.Post("/register", authHandler.Register)
			auth.Post("/verify-email", authHandler.Verify)
			auth.Post("/resend-otp", authHandler.ResendOTP)
			auth.Post("/refresh", authHandler.RefreshToken)
		})
	})

	return router
}