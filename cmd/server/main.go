package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/retail-core/auth-service/internal/api"
	"github.com/retail-core/auth-service/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	
	router := api.NewRouter()

	fmt.Println("Server running on port", cfg.PORT)
	fmt.Println("Connecting to database at", cfg.DB_SOURCE)
	err := http.ListenAndServe(":"+cfg.PORT, router)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}