package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/andro-kes/Chat/chat/internal/database"
	"github.com/andro-kes/Chat/chat/internal/handlers"
	"github.com/andro-kes/Chat/chat/internal/middlewares"
	"github.com/andro-kes/Chat/chat/logger"
	"github.com/andro-kes/Chat/chat/responses"
)

func main() {
	logger.Init()
	database.Init()
	responses.Init()

	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		log.Fatal("environment variable SECRET_KEY is required")
	}

	chatHandlers := handlers.NewChatHandlers()

	r := mux.NewRouter()

	r.Handle("/{id}/connect", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.ChatHandler)))).Methods(http.MethodGet, http.MethodPost)
	r.Handle("/{id}", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.ChatPageHandler)))).Methods(http.MethodGet)
	r.Handle("/create", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.CreateRoom)))).Methods(http.MethodPost)
	r.Handle("/{id}/messages", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.GetRoomMessages)))).Methods(http.MethodGet)
	r.Handle("/{id}/rooms", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.GetUserRooms)))).Methods(http.MethodGet)
	r.Handle("/", middlewares.RecoveryMiddleware(middlewares.AuthMiddleware(http.HandlerFunc(chatHandlers.MainPageHandler)))).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if chatHandlers != nil && chatHandlers.RabbitManager != nil {
		chatHandlers.RabbitManager.Stop()
		database.ClosePool()
	}

	log.Println("Server stopped")
}