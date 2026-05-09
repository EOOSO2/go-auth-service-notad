package server

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"auth-service/internal/application/usecase"
	"auth-service/internal/database"
	"auth-service/internal/infrastructure/auth"
	"auth-service/internal/infrastructure/auth/provider"
	redisInfra "auth-service/internal/infrastructure/cache/redis"
	"auth-service/internal/infrastructure/persistence/postgres"
	"auth-service/internal/presentation/handler"
	"auth-service/internal/presentation/middleware"
	"auth-service/internal/presentation/router"
)

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}

	dbService := database.New()
	db := dbService.GetDB()

	redisClient, err := redisInfra.NewRedisClient()
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	sessionStore := redisInfra.NewSessionStore(redisClient)

	userRepo := postgres.NewUserPostgresRepository(db)
	jwtService := auth.NewJWTService([]byte(os.Getenv("JWT_SECRET")))

	firebaseSvc, err := auth.NewFirebaseService(os.Getenv("FIREBASE_CREDENTIALS_PATH"))
	if err != nil {
		slog.Warn("Firebase unavailable, Google login disabled", "error", err)
	}

	providerFactory := provider.NewAuthProviderFactory(userRepo, firebaseSvc)
	authUC := usecase.NewAuthUseCase(userRepo, jwtService, sessionStore, providerFactory)

	authHandler := handler.NewAuthHandler(authUC)
	authMiddleware := middleware.NewAuthMiddleware(authUC)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router.NewRouter(authHandler, authMiddleware),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	slog.Info("auth service starting", "port", port)
	return srv
}
