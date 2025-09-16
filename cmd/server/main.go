package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expense-split-tracker/internal/config"
	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/middleware"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/routes"
	"expense-split-tracker/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Expense Split Tracker server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded successfully",
		zap.String("env", cfg.Server.Env),
		zap.String("db_host", cfg.Database.Host),
		zap.Int("db_port", cfg.Database.Port),
		zap.Int("server_port", cfg.Server.Port))

	// Initialize database
	db, err := database.NewConnection(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	repos := &repository.Repositories{
		User:        repository.NewUserRepository(db, logger),
		Group:       repository.NewGroupRepository(db, logger),
		Expense:     repository.NewExpenseRepository(db, logger),
		Settlement:  repository.NewSettlementRepository(db, logger),
		Balance:     repository.NewBalanceRepository(db, logger),
		Idempotency: repository.NewIdempotencyRepository(db, logger),
	}

	// Initialize services
	services := &service.Services{
		User:       service.NewUserService(repos.User, db, logger),
		Group:      service.NewGroupService(repos.Group, repos.User, db, logger),
		Expense:    service.NewExpenseService(repos.Expense, repos.Group, repos.User, repos.Balance, db, logger),
		Settlement: service.NewSettlementService(repos.Settlement, repos.Group, repos.User, repos.Balance, db, logger),
		Balance:    service.NewBalanceService(repos.Balance, repos.Group, repos.User, repos.Settlement, db, logger),
	}

	// Initialize middleware
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(repos.Idempotency, cfg, logger)
	transactionMiddleware := middleware.NewTransactionMiddleware(db, logger)

	// Start idempotency cleanup goroutine
	go idempotencyMiddleware.CleanupExpiredKeys()

	// Initialize Gin router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.StructuredLoggingMiddleware(logger))
	router.Use(gin.Recovery())
	router.Use(idempotencyMiddleware.Handle())
	router.Use(transactionMiddleware.Handle())

	// Setup routes
	routes.SetupRoutes(router, services, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server shutdown complete")
	}
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
