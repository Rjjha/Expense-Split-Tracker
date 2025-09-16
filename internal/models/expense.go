package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// SplitType represents the type of expense split
type SplitType string

const (
	SplitTypeEqual      SplitType = "equal"
	SplitTypeExact      SplitType = "exact"
	SplitTypePercentage SplitType = "percentage"
)

// Expense represents an expense in the system
type Expense struct {
	ID          int64           `json:"id" db:"id"`
	UUID        string          `json:"uuid" db:"uuid"`
	GroupID     int64           `json:"group_id" db:"group_id"`
	PaidBy      int64           `json:"paid_by" db:"paid_by"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Currency    string          `json:"currency" db:"currency"`
	Description string          `json:"description" db:"description"`
	SplitType   SplitType       `json:"split_type" db:"split_type"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// Relationships
	Group  *Group          `json:"group,omitempty"`
	Payer  *User           `json:"payer,omitempty"`
	Splits []*ExpenseSplit `json:"splits,omitempty"`
}

// ExpenseSplit represents how an expense is split among users
type ExpenseSplit struct {
	ID         int64           `json:"id" db:"id"`
	ExpenseID  int64           `json:"expense_id" db:"expense_id"`
	UserID     int64           `json:"user_id" db:"user_id"`
	Amount     decimal.Decimal `json:"amount" db:"amount"`
	Percentage decimal.Decimal `json:"percentage" db:"percentage"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`

	// Relationships
	User *User `json:"user,omitempty"`
}

// CreateExpenseRequest represents the request to create a new expense
type CreateExpenseRequest struct {
	GroupUUID   string                      `json:"group_uuid" binding:"required"`
	PaidByUUID  string                      `json:"paid_by_uuid" binding:"required"`
	Amount      decimal.Decimal             `json:"amount" binding:"required"`
	Currency    string                      `json:"currency,omitempty"`
	Description string                      `json:"description" binding:"required"`
	SplitType   SplitType                   `json:"split_type" binding:"required"`
	Splits      []CreateExpenseSplitRequest `json:"splits" binding:"required"`
}

// CreateExpenseSplitRequest represents a split in the expense creation request
type CreateExpenseSplitRequest struct {
	UserUUID   string          `json:"user_uuid" binding:"required"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	Percentage decimal.Decimal `json:"percentage,omitempty"`
}

// ExpenseListResponse represents the response for listing expenses
type ExpenseListResponse struct {
	Expenses   []*Expense `json:"expenses"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

// ExpenseFilter represents filters for expense queries
type ExpenseFilter struct {
	GroupUUID string    `json:"group_uuid,omitempty"`
	UserUUID  string    `json:"user_uuid,omitempty"`
	FromDate  time.Time `json:"from_date,omitempty"`
	ToDate    time.Time `json:"to_date,omitempty"`
	Currency  string    `json:"currency,omitempty"`
	SplitType SplitType `json:"split_type,omitempty"`
	Page      int       `json:"page,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// TableName returns the table name for Expense model
func (Expense) TableName() string {
	return "expenses"
}

// TableName returns the table name for ExpenseSplit model
func (ExpenseSplit) TableName() string {
	return "expense_splits"
}
