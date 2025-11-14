package main

import (
	"avito-pr-service/config"
	"avito-pr-service/internal/server"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.New()

	srv, err := server.New(*cfg)
	if err != nil {
		log.Fatal("failed to create server: ", err)
	}

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error: ", err)
		}
	}()

	log.Printf("Server running on :%s", cfg.Port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("shutdown error: ", err)
	}

	log.Println("Server stopped gracefully")
}
