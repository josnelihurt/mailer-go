package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedisClient(cfg config.Config) {
	if !cfg.RedisEnabled {
		return
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed (non-fatal): %v", err)
		redisClient = nil
	} else {
		log.Printf("Redis connected successfully at %s:%s", cfg.RedisHost, cfg.RedisPort)
	}
	status := redisClient.Set(ctx, "mailer-go:started", time.Now().Format(time.RFC3339), 0)
	if status.Err() != nil {
		log.Printf("Failed to set started timestamp in Redis: %v", status.Err())
	} else {
		log.Printf("Started timestamp set in Redis: %s", status.Val())
	}
}

func PushToRedis(cfg config.Config, folderName string, sms SMSMessage) {
	if !cfg.RedisEnabled || redisClient == nil {
		log.Printf("Redis is not enabled or client is not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	jsonData, err := json.Marshal(sms)
	if err != nil {
		log.Printf("Failed to marshal SMS message to JSON (non-fatal): %v", err)
		return
	}

	channelName := fmt.Sprintf("sms:%s", folderName)
	err = redisClient.Publish(ctx, channelName, string(jsonData)).Err()
	if err != nil {
		log.Printf("Failed to publish to Redis channel %s (non-fatal): %v", channelName, err)
	} else {
		log.Printf("Successfully published to Redis channel: %s", channelName)
	}
}
