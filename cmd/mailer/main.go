package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/josnelihurt/mailer-go/pkg/mailer"
)

func main() {
	log.Println("Starting mailer-go v2.1 - Native GSM Modem (no external dependencies)")

	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	log.Printf("Config loaded: modem=%s, baud=%d", cfg.ModemDevice, cfg.ModemBaud)

	// Initialize modem with native implementation
	gsmModem, err := mailer.NewGSMModem(cfg)
	if err != nil {
		log.Fatal("Failed to create modem:", err)
	}

	if err := gsmModem.Initialize(); err != nil {
		log.Fatal("Failed to initialize modem:", err)
	}

	// Initialize Redis if enabled
	mailer.InitRedisClient(cfg)

	// Start SMS polling in a separate goroutine
	go gsmModem.Start(func(sms mailer.SMSMessage) {
		log.Printf("SMS received from %s: %s", sms.From, sms.Message)

		// Send email
		if err := mailer.SendEmail(cfg, sms); err != nil {
			log.Printf("Email failed: %v", err)
		} else {
			log.Printf("Email sent successfully")
		}

		// Send to server
		if cfg.ServerURL != "" && cfg.APIKey != "" {
			client := mailer.New(cfg)
			if err := client.SendToServer("incoming", sms); err != nil {
				log.Printf("Server POST failed: %v", err)
			} else {
				log.Printf("Server POST successful")
			}
		}

		// Publish to Redis
		if cfg.RedisEnabled {
			mailer.PushToRedis(cfg, "incoming", sms)
		}

		log.Printf("SMS processed successfully")
	})

	// Configure shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
}
