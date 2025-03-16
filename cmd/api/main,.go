package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/repository"
	"github.com/dyrober/AgencyCRM/internal/server"
	"github.com/dyrober/AgencyCRM/internal/service"
)

func main() {

	//First we want to grab any config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load the configuration: %v", err)
	}

	//Now we need to connect to the DB
	db, err := repository.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to the Database: %v", err)
	}
	defer db.Close()

	//create the objects(layers) for the project
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	srv := server.NewServer(cfg, svc)

	//Start the server in a go routine
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server %v", err)
		}
		if _, err := os.Stat(cfg.TemplatesDir); os.IsNotExist(err) {
			log.Printf("Warning: Templates directory does not exist: %s", cfg.TemplatesDir)
		}
		if _, err := os.Stat(cfg.StaticDir); os.IsNotExist(err) {
			log.Printf("Warning: Static directory does not exist: %s", cfg.StaticDir)
		}
	}()

	//Wait for inpterupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	//create deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown correctly")
}
