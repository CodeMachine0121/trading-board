package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/config"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/persistence"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// shutdownAllowance lets requests in flight finish after SIGTERM before the process exits.
const shutdownAllowance = 15 * time.Second

func main() {
	if loadError := godotenv.Load(); loadError != nil {
		log.Println("no .env file loaded, falling back to process environment")
	}

	applicationConfig := config.Load()

	database, databaseError := persistence.NewDatabase(applicationConfig.Database.DataSourceName())
	if databaseError != nil {
		log.Fatalf("failed to initialize database: %v", databaseError)
	}

	engine := gin.Default()
	newRouter(database, applicationConfig).Register(engine)

	server := &http.Server{
		Addr:              ":" + applicationConfig.ServerPort,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("trading board listening on :%s", applicationConfig.ServerPort)
		if serveError := server.ListenAndServe(); serveError != nil && !errors.Is(serveError, http.ErrServerClosed) {
			log.Fatalf("server stopped: %v", serveError)
		}
	}()

	<-stopContext.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownAllowance)
	defer cancel()
	if shutdownError := server.Shutdown(shutdownContext); shutdownError != nil {
		log.Printf("graceful shutdown did not finish: %v", shutdownError)
	}
}
