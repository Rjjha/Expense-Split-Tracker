package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" db:"id"`
	UUID      string    `json:"uuid" db:"uuid"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// UserBalance represents a user's balance in a specific group
type UserBalance struct {
	UserID   int64           `json:"user_id" db:"user_id"`
	GroupID  int64           `json:"group_id" db:"group_id"`
	User     *User           `json:"user,omitempty"`
	Balance  decimal.Decimal `json:"balance" db:"balance"`
	Currency string          `json:"currency" db:"currency"`
}

// UserSummary represents a summary of user's financial status in a group
type UserSummary struct {
	User         *User           `json:"user"`
	TotalOwed    decimal.Decimal `json:"total_owed"`
	TotalOwing   decimal.Decimal `json:"total_owing"`
	NetBalance   decimal.Decimal `json:"net_balance"`
	Currency     string          `json:"currency"`
	ExpenseCount int             `json:"expense_count"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}
