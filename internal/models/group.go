package models

import (
	"time"
)

// Group represents a group in the system
type Group struct {
	ID          int64     `json:"id" db:"id"`
	UUID        string    `json:"uuid" db:"uuid"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relationships
	Creator *User   `json:"creator,omitempty"`
	Members []*User `json:"members,omitempty"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	ID       int64     `json:"id" db:"id"`
	GroupID  int64     `json:"group_id" db:"group_id"`
	UserID   int64     `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`

	// Relationships
	User *User `json:"user,omitempty"`
}

// CreateGroupRequest represents the request to create a new group
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// AddMemberRequest represents the request to add a member to a group
type AddMemberRequest struct {
	UserUUID string `json:"user_uuid" binding:"required"`
}

// GroupSummary represents a summary of group's financial status
type GroupSummary struct {
	Group        *Group         `json:"group"`
	MemberCount  int            `json:"member_count"`
	ExpenseCount int            `json:"expense_count"`
	TotalAmount  string         `json:"total_amount"`
	Currency     string         `json:"currency"`
	Balances     []*UserBalance `json:"balances,omitempty"`
}

// TableName returns the table name for Group model
func (Group) TableName() string {
	return "groups"
}

// TableName returns the table name for GroupMember model
func (GroupMember) TableName() string {
	return "group_members"
}
