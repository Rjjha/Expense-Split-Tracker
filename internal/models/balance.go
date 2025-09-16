package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Balance represents a user's balance in a group
type Balance struct {
	ID          int64           `json:"id" db:"id"`
	GroupID     int64           `json:"group_id" db:"group_id"`
	UserID      int64           `json:"user_id" db:"user_id"`
	Balance     decimal.Decimal `json:"balance" db:"balance"`
	Currency    string          `json:"currency" db:"currency"`
	LastUpdated time.Time       `json:"last_updated" db:"last_updated"`

	// Relationships
	Group *Group `json:"group,omitempty"`
	User  *User  `json:"user,omitempty"`
}

// BalanceSheet represents the complete balance sheet for a group
type BalanceSheet struct {
	Group     *Group          `json:"group"`
	Balances  []*UserBalance  `json:"balances"`
	Summary   *BalanceSummary `json:"summary"`
	Currency  string          `json:"currency"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// BalanceSummary represents summary statistics for a balance sheet
type BalanceSummary struct {
	TotalPositive decimal.Decimal `json:"total_positive"`
	TotalNegative decimal.Decimal `json:"total_negative"`
	NetBalance    decimal.Decimal `json:"net_balance"`
	UserCount     int             `json:"user_count"`
}

// UserBalanceDetail represents detailed balance information for a user
type UserBalanceDetail struct {
	User         *User             `json:"user"`
	Balance      decimal.Decimal   `json:"balance"`
	Currency     string            `json:"currency"`
	Breakdown    *BalanceBreakdown `json:"breakdown"`
	Settlements  []*Settlement     `json:"recent_settlements,omitempty"`
	LastActivity time.Time         `json:"last_activity"`
}

// BalanceBreakdown represents the breakdown of how a balance is calculated
type BalanceBreakdown struct {
	TotalPaid    decimal.Decimal `json:"total_paid"`
	TotalOwed    decimal.Decimal `json:"total_owed"`
	TotalSettled decimal.Decimal `json:"total_settled"`
	ExpenseCount int             `json:"expense_count"`
	PaymentCount int             `json:"payment_count"`
}

// DebtRelationship represents a debt relationship between two users
type DebtRelationship struct {
	Creditor *User           `json:"creditor"`
	Debtor   *User           `json:"debtor"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// TableName returns the table name for Balance model
func (Balance) TableName() string {
	return "user_balances"
}
