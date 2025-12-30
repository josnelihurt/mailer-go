package main

import (
	"fmt"
	"log"
	"os"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/josnelihurt/mailer-go/pkg/mailer"
	"github.com/josnelihurt/mailer-go/pkg/server"
)

func main() {
	fmt.Println("Starting mailer-go in server mode")
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	log.Printf("Config loaded: %v", cfg.String())

	// Check if running in server mode
	serverMode := os.Getenv("SERVER_MODE") == "true"

	if serverMode {
		// Server mode: HTTP API that receives SMS and pushes to Redis
		log.Println("Starting mailer-go in SERVER mode")

		// Initialize Redis client
		mailer.InitRedisClient(cfg)

		// Start HTTP server
		srv := server.NewServer(cfg)
		if err := srv.Start(); err != nil {
			log.Fatal("Server failed:", err)
		}
	} else {
		log.Fatal("Client mode is not supported use projects/mailer-go/cmd/mailer/main.go instead")
	}
}
