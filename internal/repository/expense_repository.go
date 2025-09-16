package repository

import (
	"context"
	"database/sql"
	"strings"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

type expenseRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewExpenseRepository creates a new expense repository
func NewExpenseRepository(db *database.DB, logger *zap.Logger) ExpenseRepository {
	return &expenseRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new expense
func (r *expenseRepository) Create(ctx context.Context, tx *database.Tx, expense *models.Expense) error {
	query := `
		INSERT INTO expenses (uuid, group_id, paid_by, amount, currency, description, split_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, expense.UUID, expense.GroupID, expense.PaidBy,
			expense.Amount, expense.Currency, expense.Description, expense.SplitType)
	} else {
		result, err = r.db.ExecContext(ctx, query, expense.UUID, expense.GroupID, expense.PaidBy,
			expense.Amount, expense.Currency, expense.Description, expense.SplitType)
	}

	if err != nil {
		r.logger.Error("Failed to create expense", zap.Error(err), zap.String("description", expense.Description))
		return errors.NewDatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	expense.ID = id
	r.logger.Info("Expense created successfully", zap.Int64("id", expense.ID), zap.String("description", expense.Description))
	return nil
}

// GetByID retrieves an expense by ID
func (r *expenseRepository) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	query := `
		SELECT e.id, e.uuid, e.group_id, e.paid_by, e.amount, e.currency, e.description, e.split_type, e.created_at, e.updated_at,
		       g.uuid as group_uuid, g.name as group_name,
		       u.uuid as payer_uuid, u.name as payer_name, u.email as payer_email
		FROM expenses e
		LEFT JOIN ` + "`groups`" + ` g ON e.group_id = g.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE e.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	expense := &models.Expense{}
	group := &models.Group{}
	payer := &models.User{}
	var groupUUID, groupName, payerUUID, payerName, payerEmail sql.NullString

	err := row.Scan(
		&expense.ID, &expense.UUID, &expense.GroupID, &expense.PaidBy, &expense.Amount,
		&expense.Currency, &expense.Description, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
		&groupUUID, &groupName,
		&payerUUID, &payerName, &payerEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Expense")
		}
		r.logger.Error("Failed to get expense by ID", zap.Error(err), zap.Int64("id", id))
		return nil, errors.NewDatabaseError(err)
	}

	if groupUUID.Valid {
		group.ID = expense.GroupID
		group.UUID = groupUUID.String
		group.Name = groupName.String
		expense.Group = group
	}

	if payerUUID.Valid {
		payer.ID = expense.PaidBy
		payer.UUID = payerUUID.String
		payer.Name = payerName.String
		payer.Email = payerEmail.String
		expense.Payer = payer
	}

	return expense, nil
}

// GetByUUID retrieves an expense by UUID
func (r *expenseRepository) GetByUUID(ctx context.Context, uuid string) (*models.Expense, error) {
	query := `
		SELECT e.id, e.uuid, e.group_id, e.paid_by, e.amount, e.currency, e.description, e.split_type, e.created_at, e.updated_at,
		       g.uuid as group_uuid, g.name as group_name,
		       u.uuid as payer_uuid, u.name as payer_name, u.email as payer_email
		FROM expenses e
		LEFT JOIN ` + "`groups`" + ` g ON e.group_id = g.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE e.uuid = ?
	`

	row := r.db.QueryRowContext(ctx, query, uuid)

	expense := &models.Expense{}
	group := &models.Group{}
	payer := &models.User{}
	var groupUUID, groupName, payerUUID, payerName, payerEmail sql.NullString

	err := row.Scan(
		&expense.ID, &expense.UUID, &expense.GroupID, &expense.PaidBy, &expense.Amount,
		&expense.Currency, &expense.Description, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
		&groupUUID, &groupName,
		&payerUUID, &payerName, &payerEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Expense")
		}
		r.logger.Error("Failed to get expense by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, errors.NewDatabaseError(err)
	}

	if groupUUID.Valid {
		group.ID = expense.GroupID
		group.UUID = groupUUID.String
		group.Name = groupName.String
		expense.Group = group
	}

	if payerUUID.Valid {
		payer.ID = expense.PaidBy
		payer.UUID = payerUUID.String
		payer.Name = payerName.String
		payer.Email = payerEmail.String
		expense.Payer = payer
	}

	return expense, nil
}

// Update updates an expense
func (r *expenseRepository) Update(ctx context.Context, tx *database.Tx, expense *models.Expense) error {
	query := `
		UPDATE expenses
		SET amount = ?, currency = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, expense.Amount, expense.Currency, expense.Description, expense.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, expense.Amount, expense.Currency, expense.Description, expense.ID)
	}

	if err != nil {
		r.logger.Error("Failed to update expense", zap.Error(err), zap.Int64("id", expense.ID))
		return errors.NewDatabaseError(err)
	}

	r.logger.Info("Expense updated successfully", zap.Int64("id", expense.ID))
	return nil
}

// Delete deletes an expense
func (r *expenseRepository) Delete(ctx context.Context, tx *database.Tx, id int64) error {
	query := `DELETE FROM expenses WHERE id = ?`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, id)
	} else {
		result, err = r.db.ExecContext(ctx, query, id)
	}

	if err != nil {
		r.logger.Error("Failed to delete expense", zap.Error(err), zap.Int64("id", id))
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Expense")
	}

	r.logger.Info("Expense deleted successfully", zap.Int64("id", id))
	return nil
}

// List retrieves expenses with filtering
func (r *expenseRepository) List(ctx context.Context, filter *models.ExpenseFilter) ([]*models.Expense, int, error) {
	whereClause := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filter.GroupUUID != "" {
		whereClause = append(whereClause, "g.uuid = ?")
		args = append(args, filter.GroupUUID)
		argIndex++
	}

	if filter.UserUUID != "" {
		whereClause = append(whereClause, "u.uuid = ?")
		args = append(args, filter.UserUUID)
		argIndex++
	}

	if filter.Currency != "" {
		whereClause = append(whereClause, "e.currency = ?")
		args = append(args, filter.Currency)
		argIndex++
	}

	if filter.SplitType != "" {
		whereClause = append(whereClause, "e.split_type = ?")
		args = append(args, filter.SplitType)
		argIndex++
	}

	if !filter.FromDate.IsZero() {
		whereClause = append(whereClause, "e.created_at >= ?")
		args = append(args, filter.FromDate)
		argIndex++
	}

	if !filter.ToDate.IsZero() {
		whereClause = append(whereClause, "e.created_at <= ?")
		args = append(args, filter.ToDate)
		argIndex++
	}

	whereSQL := strings.Join(whereClause, " AND ")

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM expenses e
		LEFT JOIN ` + "`groups`" + ` g ON e.group_id = g.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE ` + whereSQL

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count expenses", zap.Error(err))
		return nil, 0, errors.NewDatabaseError(err)
	}

	// Get data with pagination
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT e.id, e.uuid, e.group_id, e.paid_by, e.amount, e.currency, e.description, e.split_type, e.created_at, e.updated_at,
		       g.uuid as group_uuid, g.name as group_name,
		       u.uuid as payer_uuid, u.name as payer_name, u.email as payer_email
		FROM expenses e
		LEFT JOIN ` + "`groups`" + ` g ON e.group_id = g.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE ` + whereSQL + `
		ORDER BY e.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list expenses", zap.Error(err))
		return nil, 0, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		expense := &models.Expense{}
		group := &models.Group{}
		payer := &models.User{}
		var groupUUID, groupName, payerUUID, payerName, payerEmail sql.NullString

		err := rows.Scan(
			&expense.ID, &expense.UUID, &expense.GroupID, &expense.PaidBy, &expense.Amount,
			&expense.Currency, &expense.Description, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
			&groupUUID, &groupName,
			&payerUUID, &payerName, &payerEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan expense row", zap.Error(err))
			return nil, 0, errors.NewDatabaseError(err)
		}

		if groupUUID.Valid {
			group.ID = expense.GroupID
			group.UUID = groupUUID.String
			group.Name = groupName.String
			expense.Group = group
		}

		if payerUUID.Valid {
			payer.ID = expense.PaidBy
			payer.UUID = payerUUID.String
			payer.Name = payerName.String
			payer.Email = payerEmail.String
			expense.Payer = payer
		}

		expenses = append(expenses, expense)
	}

	return expenses, total, nil
}

// GetGroupExpenses retrieves expenses for a specific group
func (r *expenseRepository) GetGroupExpenses(ctx context.Context, groupID int64, offset, limit int) ([]*models.Expense, error) {
	query := `
		SELECT e.id, e.uuid, e.group_id, e.paid_by, e.amount, e.currency, e.description, e.split_type, e.created_at, e.updated_at,
		       u.uuid as payer_uuid, u.name as payer_name, u.email as payer_email
		FROM expenses e
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE e.group_id = ?
		ORDER BY e.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get group expenses", zap.Error(err), zap.Int64("groupID", groupID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		expense := &models.Expense{}
		payer := &models.User{}
		var payerUUID, payerName, payerEmail sql.NullString

		err := rows.Scan(
			&expense.ID, &expense.UUID, &expense.GroupID, &expense.PaidBy, &expense.Amount,
			&expense.Currency, &expense.Description, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
			&payerUUID, &payerName, &payerEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan group expense row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if payerUUID.Valid {
			payer.ID = expense.PaidBy
			payer.UUID = payerUUID.String
			payer.Name = payerName.String
			payer.Email = payerEmail.String
			expense.Payer = payer
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// GetUserExpenses retrieves expenses paid by a specific user
func (r *expenseRepository) GetUserExpenses(ctx context.Context, userID int64, offset, limit int) ([]*models.Expense, error) {
	query := `
		SELECT e.id, e.uuid, e.group_id, e.paid_by, e.amount, e.currency, e.description, e.split_type, e.created_at, e.updated_at,
		       g.uuid as group_uuid, g.name as group_name
		FROM expenses e
		LEFT JOIN ` + "`groups`" + ` g ON e.group_id = g.id
		WHERE e.paid_by = ?
		ORDER BY e.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get user expenses", zap.Error(err), zap.Int64("userID", userID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		expense := &models.Expense{}
		group := &models.Group{}
		var groupUUID, groupName sql.NullString

		err := rows.Scan(
			&expense.ID, &expense.UUID, &expense.GroupID, &expense.PaidBy, &expense.Amount,
			&expense.Currency, &expense.Description, &expense.SplitType, &expense.CreatedAt, &expense.UpdatedAt,
			&groupUUID, &groupName,
		)
		if err != nil {
			r.logger.Error("Failed to scan user expense row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if groupUUID.Valid {
			group.ID = expense.GroupID
			group.UUID = groupUUID.String
			group.Name = groupName.String
			expense.Group = group
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// CreateSplit creates an expense split
func (r *expenseRepository) CreateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error {
	query := `
		INSERT INTO expense_splits (expense_id, user_id, amount, percentage, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, split.ExpenseID, split.UserID, split.Amount, split.Percentage)
	} else {
		result, err = r.db.ExecContext(ctx, query, split.ExpenseID, split.UserID, split.Amount, split.Percentage)
	}

	if err != nil {
		r.logger.Error("Failed to create expense split", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	split.ID = id
	return nil
}

// GetExpenseSplits retrieves all splits for an expense
func (r *expenseRepository) GetExpenseSplits(ctx context.Context, expenseID int64) ([]*models.ExpenseSplit, error) {
	query := `
		SELECT es.id, es.expense_id, es.user_id, es.amount, es.percentage, es.created_at,
		       u.uuid, u.name, u.email
		FROM expense_splits es
		LEFT JOIN users u ON es.user_id = u.id
		WHERE es.expense_id = ?
		ORDER BY es.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, expenseID)
	if err != nil {
		r.logger.Error("Failed to get expense splits", zap.Error(err), zap.Int64("expenseID", expenseID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var splits []*models.ExpenseSplit
	for rows.Next() {
		split := &models.ExpenseSplit{}
		user := &models.User{}

		err := rows.Scan(
			&split.ID, &split.ExpenseID, &split.UserID, &split.Amount, &split.Percentage, &split.CreatedAt,
			&user.UUID, &user.Name, &user.Email,
		)
		if err != nil {
			r.logger.Error("Failed to scan expense split row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		user.ID = split.UserID
		split.User = user
		splits = append(splits, split)
	}

	return splits, nil
}

// UpdateSplit updates an expense split
func (r *expenseRepository) UpdateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error {
	query := `
		UPDATE expense_splits
		SET amount = ?, percentage = ?
		WHERE id = ?
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, split.Amount, split.Percentage, split.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, split.Amount, split.Percentage, split.ID)
	}

	if err != nil {
		r.logger.Error("Failed to update expense split", zap.Error(err), zap.Int64("id", split.ID))
		return errors.NewDatabaseError(err)
	}

	return nil
}

// DeleteExpenseSplits deletes all splits for an expense
func (r *expenseRepository) DeleteExpenseSplits(ctx context.Context, tx *database.Tx, expenseID int64) error {
	query := `DELETE FROM expense_splits WHERE expense_id = ?`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, expenseID)
	} else {
		_, err = r.db.ExecContext(ctx, query, expenseID)
	}

	if err != nil {
		r.logger.Error("Failed to delete expense splits", zap.Error(err), zap.Int64("expenseID", expenseID))
		return errors.NewDatabaseError(err)
	}

	return nil
}
