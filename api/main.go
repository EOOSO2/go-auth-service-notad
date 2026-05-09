package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"auth-service/internal/server"
)

func main() {
	srv := server.NewServer()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	// add context timeout shutdown nếu cần
}
