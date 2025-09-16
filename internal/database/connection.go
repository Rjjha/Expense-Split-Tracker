package database

import (
	"context"
	"fmt"
	"time"

	"expense-split-tracker/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// DB holds the database connection
type DB struct {
	*sqlx.DB
	logger *zap.Logger
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config, logger *zap.Logger) (*DB, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.DB.Close()
}

// BeginTx starts a new transaction
func (db *DB) BeginTx() (*Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		db.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}

	return &Tx{
		Tx:     tx,
		logger: db.logger,
	}, nil
}

// Tx represents a database transaction
type Tx struct {
	*sqlx.Tx
	logger *zap.Logger
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		tx.logger.Error("Failed to commit transaction", zap.Error(err))
	} else {
		tx.logger.Debug("Transaction committed successfully")
	}
	return err
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		tx.logger.Error("Failed to rollback transaction", zap.Error(err))
	} else {
		tx.logger.Debug("Transaction rolled back successfully")
	}
	return err
}

// WithTransaction executes a function within a database transaction
func (db *DB) WithTransaction(fn func(*Tx) error) error {
	tx, err := db.BeginTx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	err = fn(tx)
	return err
}

// Health checks the database health
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
