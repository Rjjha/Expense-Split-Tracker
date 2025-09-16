package service

import (
	"context"
	"expense-split-tracker/internal/models"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ListUsers(ctx context.Context, page, limit int) ([]*models.User, error)
}

// GroupService defines the interface for group business logic
type GroupService interface {
	CreateGroup(ctx context.Context, req *models.CreateGroupRequest, creatorUUID string) (*models.Group, error)
	GetGroupByUUID(ctx context.Context, uuid string) (*models.Group, error)
	ListGroups(ctx context.Context, page, limit int) ([]*models.Group, error)
	GetUserGroups(ctx context.Context, userUUID string, page, limit int) ([]*models.Group, error)

	// Member operations
	AddMember(ctx context.Context, groupUUID string, req *models.AddMemberRequest) error
	RemoveMember(ctx context.Context, groupUUID, userUUID string) error
	GetGroupMembers(ctx context.Context, groupUUID string) ([]*models.User, error)
}

// ExpenseService defines the interface for expense business logic
type ExpenseService interface {
	CreateExpense(ctx context.Context, req *models.CreateExpenseRequest) (*models.Expense, error)
	ListExpenses(ctx context.Context, filter *models.ExpenseFilter) (*models.ExpenseListResponse, error)
	GetGroupExpenses(ctx context.Context, groupUUID string, page, limit int) ([]*models.Expense, error)
	GetUserExpenses(ctx context.Context, userUUID string, page, limit int) ([]*models.Expense, error)
}

// SettlementService defines the interface for settlement business logic
type SettlementService interface {
	CreateSettlement(ctx context.Context, req *models.CreateSettlementRequest) (*models.Settlement, error)
	GetSettlementByUUID(ctx context.Context, uuid string) (*models.Settlement, error)
	ListSettlements(ctx context.Context, filter *models.SettlementFilter) (*models.SettlementListResponse, error)
	GetGroupSettlements(ctx context.Context, groupUUID string, page, limit int) ([]*models.Settlement, error)
	GetUserSettlements(ctx context.Context, userUUID string, page, limit int) ([]*models.Settlement, error)
	SimplifyDebts(ctx context.Context, groupUUID string) (*models.DebtSimplification, error)
}

// BalanceService defines the interface for balance business logic
type BalanceService interface {
	GetGroupBalanceSheet(ctx context.Context, groupUUID string) (*models.BalanceSheet, error)
	GetUserBalance(ctx context.Context, groupUUID, userUUID string) (*models.UserBalanceDetail, error)
	GetDebtRelationships(ctx context.Context, groupUUID string) ([]*models.DebtRelationship, error)
}

// Services aggregates all service interfaces
type Services struct {
	User       UserService
	Group      GroupService
	Expense    ExpenseService
	Settlement SettlementService
	Balance    BalanceService
}
