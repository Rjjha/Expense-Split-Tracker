package repository

import (
	"context"
	"database/sql"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/pkg/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type balanceRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewBalanceRepository creates a new balance repository
func NewBalanceRepository(db *database.DB, logger *zap.Logger) BalanceRepository {
	return &balanceRepository{
		db:     db,
		logger: logger,
	}
}

// Upsert creates or updates a balance record
func (r *balanceRepository) Upsert(ctx context.Context, tx *database.Tx, balance *models.Balance) error {
	query := `
		INSERT INTO user_balances (group_id, user_id, balance, currency, last_updated)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
		balance = VALUES(balance),
		last_updated = NOW()
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, balance.GroupID, balance.UserID, balance.Balance, balance.Currency)
	} else {
		_, err = r.db.ExecContext(ctx, query, balance.GroupID, balance.UserID, balance.Balance, balance.Currency)
	}

	if err != nil {
		r.logger.Error("Failed to upsert balance", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	return nil
}

// GetByGroupAndUser retrieves a balance for a specific group and user
func (r *balanceRepository) GetByGroupAndUser(ctx context.Context, groupID, userID int64, currency string) (*models.Balance, error) {
	query := `
		SELECT ub.id, ub.group_id, ub.user_id, ub.balance, ub.currency, ub.last_updated,
		       g.uuid as group_uuid, g.name as group_name,
		       u.uuid as user_uuid, u.name as user_name, u.email as user_email
		FROM user_balances ub
		LEFT JOIN ` + "`groups`" + ` g ON ub.group_id = g.id
		LEFT JOIN users u ON ub.user_id = u.id
		WHERE ub.group_id = ? AND ub.user_id = ? AND ub.currency = ?
	`

	balance := &models.Balance{}
	group := &models.Group{}
	user := &models.User{}
	var groupUUID, groupName, userUUID, userName, userEmail sql.NullString

	err := r.db.QueryRowContext(ctx, query, groupID, userID, currency).Scan(
		&balance.ID, &balance.GroupID, &balance.UserID, &balance.Balance, &balance.Currency, &balance.LastUpdated,
		&groupUUID, &groupName,
		&userUUID, &userName, &userEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return zero balance if no record exists
			balance.GroupID = groupID
			balance.UserID = userID
			balance.Balance = decimal.Zero
			balance.Currency = currency
			return balance, nil
		}
		r.logger.Error("Failed to get balance", zap.Error(err))
		return nil, errors.NewDatabaseError(err)
	}

	if groupUUID.Valid {
		group.ID = balance.GroupID
		group.UUID = groupUUID.String
		group.Name = groupName.String
		balance.Group = group
	}

	if userUUID.Valid {
		user.ID = balance.UserID
		user.UUID = userUUID.String
		user.Name = userName.String
		user.Email = userEmail.String
		balance.User = user
	}

	return balance, nil
}

// GetGroupBalances retrieves all balances for a group
func (r *balanceRepository) GetGroupBalances(ctx context.Context, groupID int64, currency string) ([]*models.Balance, error) {
	query := `
		SELECT ub.id, ub.group_id, ub.user_id, ub.balance, ub.currency, ub.last_updated,
		       u.uuid as user_uuid, u.name as user_name, u.email as user_email
		FROM user_balances ub
		LEFT JOIN users u ON ub.user_id = u.id
		WHERE ub.group_id = ? AND ub.currency = ?
		ORDER BY ub.balance DESC
	`

	rows, err := r.db.QueryContext(ctx, query, groupID, currency)
	if err != nil {
		r.logger.Error("Failed to get group balances", zap.Error(err))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var balances []*models.Balance
	for rows.Next() {
		balance := &models.Balance{}
		user := &models.User{}
		var userUUID, userName, userEmail sql.NullString

		err := rows.Scan(
			&balance.ID, &balance.GroupID, &balance.UserID, &balance.Balance, &balance.Currency, &balance.LastUpdated,
			&userUUID, &userName, &userEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan balance row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if userUUID.Valid {
			user.ID = balance.UserID
			user.UUID = userUUID.String
			user.Name = userName.String
			user.Email = userEmail.String
			balance.User = user
		}

		balances = append(balances, balance)
	}

	return balances, nil
}

// GetUserBalances retrieves all balances for a user across all groups
func (r *balanceRepository) GetUserBalances(ctx context.Context, userID int64) ([]*models.Balance, error) {
	query := `
		SELECT ub.id, ub.group_id, ub.user_id, ub.balance, ub.currency, ub.last_updated,
		       g.uuid as group_uuid, g.name as group_name
		FROM user_balances ub
		LEFT JOIN ` + "`groups`" + ` g ON ub.group_id = g.id
		WHERE ub.user_id = ?
		ORDER BY ub.last_updated DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get user balances", zap.Error(err))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var balances []*models.Balance
	for rows.Next() {
		balance := &models.Balance{}
		group := &models.Group{}
		var groupUUID, groupName sql.NullString

		err := rows.Scan(
			&balance.ID, &balance.GroupID, &balance.UserID, &balance.Balance, &balance.Currency, &balance.LastUpdated,
			&groupUUID, &groupName,
		)
		if err != nil {
			r.logger.Error("Failed to scan user balance row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if groupUUID.Valid {
			group.ID = balance.GroupID
			group.UUID = groupUUID.String
			group.Name = groupName.String
			balance.Group = group
		}

		balances = append(balances, balance)
	}

	return balances, nil
}

// UpdateBalance updates a user's balance by adding/subtracting an amount
func (r *balanceRepository) UpdateBalance(ctx context.Context, tx *database.Tx, groupID, userID int64, amount decimal.Decimal, currency string) error {
	query := `
		INSERT INTO user_balances (group_id, user_id, balance, currency, last_updated)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
		balance = balance + VALUES(balance),
		last_updated = NOW()
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, groupID, userID, amount, currency)
	} else {
		_, err = r.db.ExecContext(ctx, query, groupID, userID, amount, currency)
	}

	if err != nil {
		r.logger.Error("Failed to update balance", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	return nil
}
