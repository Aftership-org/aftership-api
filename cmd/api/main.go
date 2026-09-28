package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Aftership-org/aftership-api/internal/config"
	"github.com/Aftership-org/aftership-api/internal/db"
	"github.com/Aftership-org/aftership-api/internal/httpx"
	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	// Load configuration and initialize database connection
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.NewPostgres(cfg.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal(err)
	}

	// Health check endpoint
	router.Get("/health", healthCheck)

	address := ":" + cfg.AppPort
	server := &http.Server{
		Addr:    address,
		Handler: router,
	}
	// Graceful shutdown
	go func() {
		log.Printf("server listening on http://localhost:%s", cfg.AppPort)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Wait here until the process receives Ctrl+C or SIGTERM.
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("database close error: %v", err)
	}

	log.Println("server stopped")
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"service": "aftership-api",
		"version": "1.0.0",
	}
	httpx.JSON(w, http.StatusOK, response)
}
