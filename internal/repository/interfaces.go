package repository

import (
	"context"
	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"

	"github.com/shopspring/decimal"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, tx *database.Tx, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
}

// GroupRepository defines the interface for group data operations
type GroupRepository interface {
	Create(ctx context.Context, tx *database.Tx, group *models.Group) error
	GetByID(ctx context.Context, id int64) (*models.Group, error)
	GetByUUID(ctx context.Context, uuid string) (*models.Group, error)
	List(ctx context.Context, offset, limit int) ([]*models.Group, error)
	GetUserGroups(ctx context.Context, userID int64, offset, limit int) ([]*models.Group, error)

	// Member operations
	AddMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error
	RemoveMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error
	GetMembers(ctx context.Context, groupID int64) ([]*models.User, error)
	IsMember(ctx context.Context, groupID, userID int64) (bool, error)
}

// ExpenseRepository defines the interface for expense data operations
type ExpenseRepository interface {
	Create(ctx context.Context, tx *database.Tx, expense *models.Expense) error
	GetByID(ctx context.Context, id int64) (*models.Expense, error)
	List(ctx context.Context, filter *models.ExpenseFilter) ([]*models.Expense, int, error)
	GetGroupExpenses(ctx context.Context, groupID int64, offset, limit int) ([]*models.Expense, error)
	GetUserExpenses(ctx context.Context, userID int64, offset, limit int) ([]*models.Expense, error)

	// Split operations
	CreateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error
	GetExpenseSplits(ctx context.Context, expenseID int64) ([]*models.ExpenseSplit, error)
	UpdateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error
}

// SettlementRepository defines the interface for settlement data operations
type SettlementRepository interface {
	Create(ctx context.Context, tx *database.Tx, settlement *models.Settlement) error
	GetByID(ctx context.Context, id int64) (*models.Settlement, error)
	GetByUUID(ctx context.Context, uuid string) (*models.Settlement, error)
	List(ctx context.Context, filter *models.SettlementFilter) ([]*models.Settlement, int, error)
	GetGroupSettlements(ctx context.Context, groupID int64, offset, limit int) ([]*models.Settlement, error)
	GetUserSettlements(ctx context.Context, userID int64, offset, limit int) ([]*models.Settlement, error)
}

// BalanceRepository defines the interface for balance data operations
type BalanceRepository interface {
	Upsert(ctx context.Context, tx *database.Tx, balance *models.Balance) error
	GetByGroupAndUser(ctx context.Context, groupID, userID int64, currency string) (*models.Balance, error)
	GetGroupBalances(ctx context.Context, groupID int64, currency string) ([]*models.Balance, error)
	GetUserBalances(ctx context.Context, userID int64) ([]*models.Balance, error)
	UpdateBalance(ctx context.Context, tx *database.Tx, groupID, userID int64, amount decimal.Decimal, currency string) error
}

// IdempotencyRepository defines the interface for idempotency key operations
type IdempotencyRepository interface {
	Create(ctx context.Context, tx *database.Tx, key, requestHash string, responseData []byte, statusCode int, expiresAt int64) error
	GetByKey(ctx context.Context, key string) (*IdempotencyRecord, error)
	DeleteExpired(ctx context.Context, tx *database.Tx) error
}

// IdempotencyRecord represents an idempotency record
type IdempotencyRecord struct {
	ID           int64  `json:"id" db:"id"`
	KeyValue     string `json:"key_value" db:"key_value"`
	RequestHash  string `json:"request_hash" db:"request_hash"`
	ResponseData []byte `json:"response_data" db:"response_data"`
	StatusCode   int    `json:"status_code" db:"status_code"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	ExpiresAt    int64  `json:"expires_at" db:"expires_at"`
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	User        UserRepository
	Group       GroupRepository
	Expense     ExpenseRepository
	Settlement  SettlementRepository
	Balance     BalanceRepository
	Idempotency IdempotencyRepository
}
