package service

import (
	"context"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/utils"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

// DBTransactor defines the interface for database transaction operations
type DBTransactor interface {
	WithTransaction(fn func(*database.Tx) error) error
}

type userService struct {
	repo   repository.UserRepository
	db     DBTransactor
	logger *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, db DBTransactor, logger *zap.Logger) UserService {
	return &userService{
		repo:   repo,
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Validate input
	if err := utils.ValidateName(req.Name); err != nil {
		return nil, err
	}

	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	// Check if user with email already exists
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.NewAlreadyExistsError("User with this email")
	}

	// If error is not "not found", return it
	if err != nil {
		if appErr, ok := err.(*errors.AppError); !ok || appErr.Code != errors.ErrCodeNotFound {
			return nil, err
		}
	}

	// Create user with transaction
	user := &models.User{
		UUID:  utils.GenerateUUID(),
		Name:  req.Name,
		Email: req.Email,
	}

	err = s.db.WithTransaction(func(tx *database.Tx) error {
		return s.repo.Create(ctx, tx, user)
	})

	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, err
	}

	s.logger.Info("User created successfully", zap.String("uuid", user.UUID), zap.String("email", user.Email))
	return user, nil
}

// GetUserByUUID retrieves a user by UUID
func (s *userService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	if !utils.IsValidUUID(uuid) {
		return nil, errors.NewInvalidValueError("uuid", uuid)
	}

	user, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		s.logger.Error("Failed to get user by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if err := utils.ValidateEmail(email); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	return user, nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]*models.User, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, err
	}

	return users, nil
}
