package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/M1ralai/notly-api/internal/app"
	"github.com/M1ralai/notly-api/internal/infrastructure/database"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/infrastructure/metrics"
	"github.com/joho/godotenv"

	_ "github.com/M1ralai/notly-api/docs"
)

// @title Notly API
// @version 1.0
// @description Notly projesinin backend API dokümantasyonu.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("⚠ .env file not found, using environment variables")
	}

	metrics.Init()
	db := database.NewDb()
	zapLogger := logger.NewLogger(db.Conn)

	log.Println("✓ Running database migrations...")
	if err := database.RunMigrations(db.Conn.DB); err != nil {
		log.Fatalf("✗ Migration failed: %v", err)
	}
	log.Println("✓ Migrations completed successfully")

	server := app.NewServer(db.Conn, zapLogger)

	if err := server.Start(); err != nil {
		log.Fatalf("✗ Failed to start server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("\n✓ Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("✗ Shutdown failed: %v", err)
	}

	log.Println("✓ Server exited cleanly")
}
