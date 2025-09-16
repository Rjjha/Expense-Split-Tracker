package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Settlement represents a debt settlement between users
type Settlement struct {
	ID          int64           `json:"id" db:"id"`
	UUID        string          `json:"uuid" db:"uuid"`
	GroupID     int64           `json:"group_id" db:"group_id"`
	FromUserID  int64           `json:"from_user_id" db:"from_user_id"`
	ToUserID    int64           `json:"to_user_id" db:"to_user_id"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Currency    string          `json:"currency" db:"currency"`
	Description string          `json:"description" db:"description"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`

	// Relationships
	Group    *Group `json:"group,omitempty"`
	FromUser *User  `json:"from_user,omitempty"`
	ToUser   *User  `json:"to_user,omitempty"`
}

// CreateSettlementRequest represents the request to create a new settlement
type CreateSettlementRequest struct {
	GroupUUID    string          `json:"group_uuid" binding:"required"`
	FromUserUUID string          `json:"from_user_uuid" binding:"required"`
	ToUserUUID   string          `json:"to_user_uuid" binding:"required"`
	Amount       decimal.Decimal `json:"amount" binding:"required"`
	Currency     string          `json:"currency,omitempty"`
	Description  string          `json:"description,omitempty"`
}

// SettlementSuggestion represents a suggested settlement to simplify debts
type SettlementSuggestion struct {
	FromUser *User           `json:"from_user"`
	ToUser   *User           `json:"to_user"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// DebtSimplification represents the result of debt simplification
type DebtSimplification struct {
	OriginalTransactions   int                     `json:"original_transactions"`
	SimplifiedTransactions int                     `json:"simplified_transactions"`
	Savings                int                     `json:"savings"`
	Suggestions            []*SettlementSuggestion `json:"suggestions"`
}

// SettlementListResponse represents the response for listing settlements
type SettlementListResponse struct {
	Settlements []*Settlement `json:"settlements"`
	TotalCount  int           `json:"total_count"`
	Page        int           `json:"page"`
	Limit       int           `json:"limit"`
}

// SettlementFilter represents filters for settlement queries
type SettlementFilter struct {
	GroupUUID    string    `json:"group_uuid,omitempty"`
	UserUUID     string    `json:"user_uuid,omitempty"`
	FromUserUUID string    `json:"from_user_uuid,omitempty"`
	ToUserUUID   string    `json:"to_user_uuid,omitempty"`
	FromDate     time.Time `json:"from_date,omitempty"`
	ToDate       time.Time `json:"to_date,omitempty"`
	Currency     string    `json:"currency,omitempty"`
	Page         int       `json:"page,omitempty"`
	Limit        int       `json:"limit,omitempty"`
}

// TableName returns the table name for Settlement model
func (Settlement) TableName() string {
	return "settlements"
}
