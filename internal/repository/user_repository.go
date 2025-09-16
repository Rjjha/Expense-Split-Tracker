package repository

import (
	"context"
	"database/sql"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

type userRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, tx *database.Tx, user *models.User) error {
	query := `
		INSERT INTO users (uuid, name, email, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, user.UUID, user.Name, user.Email)
	} else {
		result, err = r.db.ExecContext(ctx, query, user.UUID, user.Name, user.Email)
	}

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("email", user.Email))
		return errors.NewDatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	user.ID = id
	r.logger.Info("User created successfully", zap.Int64("id", user.ID), zap.String("email", user.Email))
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, uuid, name, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("User")
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.Int64("id", id))
		return nil, errors.NewDatabaseError(err)
	}

	return user, nil
}

// GetByUUID retrieves a user by UUID
func (r *userRepository) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	query := `
		SELECT id, uuid, name, email, created_at, updated_at
		FROM users
		WHERE uuid = ?
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("User")
		}
		r.logger.Error("Failed to get user by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, errors.NewDatabaseError(err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, uuid, name, email, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("User")
		}
		r.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, errors.NewDatabaseError(err)
	}

	return user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, tx *database.Tx, user *models.User) error {
	query := `
		UPDATE users
		SET name = ?, email = ?, updated_at = NOW()
		WHERE id = ?
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	}

	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.Int64("id", user.ID))
		return errors.NewDatabaseError(err)
	}

	r.logger.Info("User updated successfully", zap.Int64("id", user.ID))
	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, tx *database.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, id)
	} else {
		result, err = r.db.ExecContext(ctx, query, id)
	}

	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.Int64("id", id))
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("User")
	}

	r.logger.Info("User deleted successfully", zap.Int64("id", id))
	return nil
}

// List retrieves a list of users with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	query := `
		SELECT id, uuid, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	users := []*models.User{}
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, errors.NewDatabaseError(err)
	}

	return users, nil
}
