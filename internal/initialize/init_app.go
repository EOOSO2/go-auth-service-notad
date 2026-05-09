package initialize

import (
	"auth-service/internal/database"
	"auth-service/internal/infrastructure/cache/redis"
	"log"
	"os"
)

func InitApp() {
	// Init DB
	db := database.New()
	err := db.GetDB().Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Init Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL not set")
	}
	_, err = redis.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("App initialized successfully")
}
