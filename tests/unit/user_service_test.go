package unit

import (
	"context"
	"testing"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"
	"expense-split-tracker/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, tx *database.Tx, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, tx *database.Tx, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, tx *database.Tx, id int64) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockDB is a mock implementation of service.DBTransactor
type MockDB struct {
	mock.Mock
}

func (m *MockDB) WithTransaction(fn func(tx *database.Tx) error) error {
	args := m.Called(fn)
	// Execute the function with nil transaction for testing
	if err := fn(nil); err != nil {
		return err
	}
	return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateUserRequest
		setupMocks    func(*MockUserRepository, *MockDB)
		expectedError string
		expectedUser  *models.User
	}{
		{
			name: "successful user creation",
			request: &models.CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			setupMocks: func(repo *MockUserRepository, db *MockDB) {
				// Mock email check - user doesn't exist
				repo.On("GetByEmail", mock.Anything, "john@example.com").
					Return(nil, errors.NewNotFoundError("User"))

				// Mock user creation
				repo.On("Create", mock.Anything, (*database.Tx)(nil), mock.MatchedBy(func(u *models.User) bool {
					return u.Name == "John Doe" && u.Email == "john@example.com" && u.UUID != ""
				})).Return(nil)

				// Mock transaction
				db.On("WithTransaction", mock.AnythingOfType("func(*database.Tx) error")).Return(nil)
			},
			expectedUser: &models.User{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name: "user already exists",
			request: &models.CreateUserRequest{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			setupMocks: func(repo *MockUserRepository, db *MockDB) {
				// Mock email check - user exists
				existingUser := &models.User{
					ID:    1,
					Name:  "Jane Doe",
					Email: "jane@example.com",
				}
				repo.On("GetByEmail", mock.Anything, "jane@example.com").
					Return(existingUser, nil)
			},
			expectedError: "User with this email already exists",
		},
		{
			name: "invalid email",
			request: &models.CreateUserRequest{
				Name:  "Invalid User",
				Email: "invalid-email",
			},
			setupMocks:    func(repo *MockUserRepository, db *MockDB) {},
			expectedError: "Invalid value 'invalid-email' for field 'email'",
		},
		{
			name: "empty name",
			request: &models.CreateUserRequest{
				Name:  "",
				Email: "test@example.com",
			},
			setupMocks:    func(repo *MockUserRepository, db *MockDB) {},
			expectedError: "Field 'name' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockUserRepository)
			mockDB := new(MockDB)
			logger := zaptest.NewLogger(t)

			tt.setupMocks(mockRepo, mockDB)

			// Create service
			userService := service.NewUserService(mockRepo, mockDB, logger)

			// Execute test
			result, err := userService.CreateUser(context.Background(), tt.request)

			// Verify results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.Name, result.Name)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.NotEmpty(t, result.UUID)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByUUID(t *testing.T) {
	tests := []struct {
		name          string
		uuid          string
		setupMocks    func(*MockUserRepository)
		expectedError string
		expectedUser  *models.User
	}{
		{
			name: "successful user retrieval",
			uuid: "550e8400-e29b-41d4-a716-446655440000",
			setupMocks: func(repo *MockUserRepository) {
				user := &models.User{
					ID:    1,
					UUID:  "550e8400-e29b-41d4-a716-446655440000",
					Name:  "John Doe",
					Email: "john@example.com",
				}
				repo.On("GetByUUID", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").
					Return(user, nil)
			},
			expectedUser: &models.User{
				ID:    1,
				UUID:  "550e8400-e29b-41d4-a716-446655440000",
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name: "user not found",
			uuid: "550e8400-e29b-41d4-a716-446655440000",
			setupMocks: func(repo *MockUserRepository) {
				repo.On("GetByUUID", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").
					Return(nil, errors.NewNotFoundError("User"))
			},
			expectedError: "User not found",
		},
		{
			name:          "invalid uuid",
			uuid:          "invalid-uuid",
			setupMocks:    func(repo *MockUserRepository) {},
			expectedError: "Invalid value 'invalid-uuid' for field 'uuid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockUserRepository)
			mockDB := new(MockDB)
			logger := zaptest.NewLogger(t)

			tt.setupMocks(mockRepo)

			// Create service
			userService := service.NewUserService(mockRepo, mockDB, logger)

			// Execute test
			result, err := userService.GetUserByUUID(context.Background(), tt.uuid)

			// Verify results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.UUID, result.UUID)
				assert.Equal(t, tt.expectedUser.Name, result.Name)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
