package repository

import (
	"context"
	"database/sql"
	"time"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

type idempotencyRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewIdempotencyRepository creates a new idempotency repository
func NewIdempotencyRepository(db *database.DB, logger *zap.Logger) IdempotencyRepository {
	return &idempotencyRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new idempotency record
func (r *idempotencyRepository) Create(ctx context.Context, tx *database.Tx, key, requestHash string, responseData []byte, statusCode int, expiresAt int64) error {
	query := `
		INSERT INTO idempotency_keys (key_value, request_hash, response_data, status_code, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	createdAt := time.Now().Unix()

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, key, requestHash, responseData, statusCode, createdAt, expiresAt)
	} else {
		_, err = r.db.ExecContext(ctx, query, key, requestHash, responseData, statusCode, createdAt, expiresAt)
	}

	if err != nil {
		r.logger.Error("Failed to create idempotency record", zap.Error(err), zap.String("key", key))
		return errors.NewDatabaseError(err)
	}

	r.logger.Debug("Idempotency record created successfully", zap.String("key", key))
	return nil
}

// GetByKey retrieves an idempotency record by key
func (r *idempotencyRepository) GetByKey(ctx context.Context, key string) (*IdempotencyRecord, error) {
	query := `
		SELECT id, key_value, request_hash, response_data, status_code, created_at, expires_at
		FROM idempotency_keys
		WHERE key_value = ? AND expires_at > ?
	`

	now := time.Now().Unix()
	record := &IdempotencyRecord{}

	err := r.db.GetContext(ctx, record, query, key, now)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, but not an error
		}
		r.logger.Error("Failed to get idempotency record", zap.Error(err), zap.String("key", key))
		return nil, errors.NewDatabaseError(err)
	}

	return record, nil
}

// DeleteExpired deletes expired idempotency records
func (r *idempotencyRepository) DeleteExpired(ctx context.Context, tx *database.Tx) error {
	query := `DELETE FROM idempotency_keys WHERE expires_at <= ?`

	now := time.Now().Unix()

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, now)
	} else {
		result, err = r.db.ExecContext(ctx, query, now)
	}

	if err != nil {
		r.logger.Error("Failed to delete expired idempotency records", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected > 0 {
		r.logger.Info("Deleted expired idempotency records", zap.Int64("count", rowsAffected))
	}

	return nil
}
