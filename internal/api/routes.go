package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/db"
	"github.com/retail-core/auth-service/internal/user"
	"github.com/retail-core/auth-service/internal/middleware"
)

func NewRouter() http.Handler {
	cfg := config.LoadConfig()
	router := chi.NewRouter()

	db, query := db.Connect(cfg.DB_SOURCE)
	userRepo := user.NewPGRepository(query, db)

	authService := auth.NewService(cfg.JWT_SECRET_KEY, userRepo)
	authHandler := NewAuthHandler(authService)

	router.Use(middleware.RequestLogger)

	router.Get("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ok"))
	})

	router.Route("/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(auth chi.Router) {
			auth.Post("/login", authHandler.Login)
			auth.Post("/register", authHandler.Register)
			auth.Post("/verify-email", authHandler.Verify)
			auth.Post("/resend-otp", authHandler.ResendOTP)
			auth.Post("/refresh", authHandler.RefreshToken)
		})
	})

	return router
}