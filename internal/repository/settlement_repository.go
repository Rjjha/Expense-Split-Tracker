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

type settlementRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewSettlementRepository creates a new settlement repository
func NewSettlementRepository(db *database.DB, logger *zap.Logger) SettlementRepository {
	return &settlementRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new settlement
func (r *settlementRepository) Create(ctx context.Context, tx *database.Tx, settlement *models.Settlement) error {
	query := `
		INSERT INTO settlements (uuid, group_id, from_user_id, to_user_id, amount, currency, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, settlement.UUID, settlement.GroupID, settlement.FromUserID,
			settlement.ToUserID, settlement.Amount, settlement.Currency, settlement.Description)
	} else {
		result, err = r.db.ExecContext(ctx, query, settlement.UUID, settlement.GroupID, settlement.FromUserID,
			settlement.ToUserID, settlement.Amount, settlement.Currency, settlement.Description)
	}

	if err != nil {
		r.logger.Error("Failed to create settlement", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	settlement.ID = id
	r.logger.Info("Settlement created successfully", zap.Int64("id", settlement.ID))
	return nil
}

// GetByID retrieves a settlement by ID
func (r *settlementRepository) GetByID(ctx context.Context, id int64) (*models.Settlement, error) {
	query := `
		SELECT s.id, s.uuid, s.group_id, s.from_user_id, s.to_user_id, s.amount, s.currency, s.description, s.created_at,
		       g.uuid as group_uuid, g.name as group_name,
		       fu.uuid as from_user_uuid, fu.name as from_user_name, fu.email as from_user_email,
		       tu.uuid as to_user_uuid, tu.name as to_user_name, tu.email as to_user_email
		FROM settlements s
		LEFT JOIN ` + "`groups`" + ` g ON s.group_id = g.id
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE s.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	settlement := &models.Settlement{}
	group := &models.Group{}
	fromUser := &models.User{}
	toUser := &models.User{}
	var groupUUID, groupName, fromUserUUID, fromUserName, fromUserEmail, toUserUUID, toUserName, toUserEmail sql.NullString

	err := row.Scan(
		&settlement.ID, &settlement.UUID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
		&settlement.Amount, &settlement.Currency, &settlement.Description, &settlement.CreatedAt,
		&groupUUID, &groupName,
		&fromUserUUID, &fromUserName, &fromUserEmail,
		&toUserUUID, &toUserName, &toUserEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Settlement")
		}
		r.logger.Error("Failed to get settlement by ID", zap.Error(err), zap.Int64("id", id))
		return nil, errors.NewDatabaseError(err)
	}

	if groupUUID.Valid {
		group.ID = settlement.GroupID
		group.UUID = groupUUID.String
		group.Name = groupName.String
		settlement.Group = group
	}

	if fromUserUUID.Valid {
		fromUser.ID = settlement.FromUserID
		fromUser.UUID = fromUserUUID.String
		fromUser.Name = fromUserName.String
		fromUser.Email = fromUserEmail.String
		settlement.FromUser = fromUser
	}

	if toUserUUID.Valid {
		toUser.ID = settlement.ToUserID
		toUser.UUID = toUserUUID.String
		toUser.Name = toUserName.String
		toUser.Email = toUserEmail.String
		settlement.ToUser = toUser
	}

	return settlement, nil
}

// GetByUUID retrieves a settlement by UUID
func (r *settlementRepository) GetByUUID(ctx context.Context, uuid string) (*models.Settlement, error) {
	query := `
		SELECT s.id, s.uuid, s.group_id, s.from_user_id, s.to_user_id, s.amount, s.currency, s.description, s.created_at,
		       g.uuid as group_uuid, g.name as group_name,
		       fu.uuid as from_user_uuid, fu.name as from_user_name, fu.email as from_user_email,
		       tu.uuid as to_user_uuid, tu.name as to_user_name, tu.email as to_user_email
		FROM settlements s
		LEFT JOIN ` + "`groups`" + ` g ON s.group_id = g.id
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE s.uuid = ?
	`

	row := r.db.QueryRowContext(ctx, query, uuid)

	settlement := &models.Settlement{}
	group := &models.Group{}
	fromUser := &models.User{}
	toUser := &models.User{}
	var groupUUID, groupName, fromUserUUID, fromUserName, fromUserEmail, toUserUUID, toUserName, toUserEmail sql.NullString

	err := row.Scan(
		&settlement.ID, &settlement.UUID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
		&settlement.Amount, &settlement.Currency, &settlement.Description, &settlement.CreatedAt,
		&groupUUID, &groupName,
		&fromUserUUID, &fromUserName, &fromUserEmail,
		&toUserUUID, &toUserName, &toUserEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Settlement")
		}
		r.logger.Error("Failed to get settlement by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, errors.NewDatabaseError(err)
	}

	if groupUUID.Valid {
		group.ID = settlement.GroupID
		group.UUID = groupUUID.String
		group.Name = groupName.String
		settlement.Group = group
	}

	if fromUserUUID.Valid {
		fromUser.ID = settlement.FromUserID
		fromUser.UUID = fromUserUUID.String
		fromUser.Name = fromUserName.String
		fromUser.Email = fromUserEmail.String
		settlement.FromUser = fromUser
	}

	if toUserUUID.Valid {
		toUser.ID = settlement.ToUserID
		toUser.UUID = toUserUUID.String
		toUser.Name = toUserName.String
		toUser.Email = toUserEmail.String
		settlement.ToUser = toUser
	}

	return settlement, nil
}

// List retrieves settlements with filtering
func (r *settlementRepository) List(ctx context.Context, filter *models.SettlementFilter) ([]*models.Settlement, int, error) {
	whereClause := []string{"1=1"}
	args := []interface{}{}

	if filter.GroupUUID != "" {
		whereClause = append(whereClause, "g.uuid = ?")
		args = append(args, filter.GroupUUID)
	}

	if filter.UserUUID != "" {
		whereClause = append(whereClause, "(fu.uuid = ? OR tu.uuid = ?)")
		args = append(args, filter.UserUUID, filter.UserUUID)
	}

	if filter.FromUserUUID != "" {
		whereClause = append(whereClause, "fu.uuid = ?")
		args = append(args, filter.FromUserUUID)
	}

	if filter.ToUserUUID != "" {
		whereClause = append(whereClause, "tu.uuid = ?")
		args = append(args, filter.ToUserUUID)
	}

	if filter.Currency != "" {
		whereClause = append(whereClause, "s.currency = ?")
		args = append(args, filter.Currency)
	}

	if !filter.FromDate.IsZero() {
		whereClause = append(whereClause, "s.created_at >= ?")
		args = append(args, filter.FromDate)
	}

	if !filter.ToDate.IsZero() {
		whereClause = append(whereClause, "s.created_at <= ?")
		args = append(args, filter.ToDate)
	}

	whereSQL := strings.Join(whereClause, " AND ")

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM settlements s
		LEFT JOIN ` + "`groups`" + ` g ON s.group_id = g.id
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE ` + whereSQL

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count settlements", zap.Error(err))
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
		SELECT s.id, s.uuid, s.group_id, s.from_user_id, s.to_user_id, s.amount, s.currency, s.description, s.created_at,
		       g.uuid as group_uuid, g.name as group_name,
		       fu.uuid as from_user_uuid, fu.name as from_user_name, fu.email as from_user_email,
		       tu.uuid as to_user_uuid, tu.name as to_user_name, tu.email as to_user_email
		FROM settlements s
		LEFT JOIN ` + "`groups`" + ` g ON s.group_id = g.id
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE ` + whereSQL + `
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list settlements", zap.Error(err))
		return nil, 0, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var settlements []*models.Settlement
	for rows.Next() {
		settlement := &models.Settlement{}
		group := &models.Group{}
		fromUser := &models.User{}
		toUser := &models.User{}
		var groupUUID, groupName, fromUserUUID, fromUserName, fromUserEmail, toUserUUID, toUserName, toUserEmail sql.NullString

		err := rows.Scan(
			&settlement.ID, &settlement.UUID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
			&settlement.Amount, &settlement.Currency, &settlement.Description, &settlement.CreatedAt,
			&groupUUID, &groupName,
			&fromUserUUID, &fromUserName, &fromUserEmail,
			&toUserUUID, &toUserName, &toUserEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan settlement row", zap.Error(err))
			return nil, 0, errors.NewDatabaseError(err)
		}

		if groupUUID.Valid {
			group.ID = settlement.GroupID
			group.UUID = groupUUID.String
			group.Name = groupName.String
			settlement.Group = group
		}

		if fromUserUUID.Valid {
			fromUser.ID = settlement.FromUserID
			fromUser.UUID = fromUserUUID.String
			fromUser.Name = fromUserName.String
			fromUser.Email = fromUserEmail.String
			settlement.FromUser = fromUser
		}

		if toUserUUID.Valid {
			toUser.ID = settlement.ToUserID
			toUser.UUID = toUserUUID.String
			toUser.Name = toUserName.String
			toUser.Email = toUserEmail.String
			settlement.ToUser = toUser
		}

		settlements = append(settlements, settlement)
	}

	return settlements, total, nil
}

// GetGroupSettlements retrieves settlements for a specific group
func (r *settlementRepository) GetGroupSettlements(ctx context.Context, groupID int64, offset, limit int) ([]*models.Settlement, error) {
	query := `
		SELECT s.id, s.uuid, s.group_id, s.from_user_id, s.to_user_id, s.amount, s.currency, s.description, s.created_at,
		       fu.uuid as from_user_uuid, fu.name as from_user_name, fu.email as from_user_email,
		       tu.uuid as to_user_uuid, tu.name as to_user_name, tu.email as to_user_email
		FROM settlements s
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE s.group_id = ?
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get group settlements", zap.Error(err), zap.Int64("groupID", groupID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var settlements []*models.Settlement
	for rows.Next() {
		settlement := &models.Settlement{}
		fromUser := &models.User{}
		toUser := &models.User{}
		var fromUserUUID, fromUserName, fromUserEmail, toUserUUID, toUserName, toUserEmail sql.NullString

		err := rows.Scan(
			&settlement.ID, &settlement.UUID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
			&settlement.Amount, &settlement.Currency, &settlement.Description, &settlement.CreatedAt,
			&fromUserUUID, &fromUserName, &fromUserEmail,
			&toUserUUID, &toUserName, &toUserEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan group settlement row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if fromUserUUID.Valid {
			fromUser.ID = settlement.FromUserID
			fromUser.UUID = fromUserUUID.String
			fromUser.Name = fromUserName.String
			fromUser.Email = fromUserEmail.String
			settlement.FromUser = fromUser
		}

		if toUserUUID.Valid {
			toUser.ID = settlement.ToUserID
			toUser.UUID = toUserUUID.String
			toUser.Name = toUserName.String
			toUser.Email = toUserEmail.String
			settlement.ToUser = toUser
		}

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

// GetUserSettlements retrieves settlements for a specific user (either as payer or receiver)
func (r *settlementRepository) GetUserSettlements(ctx context.Context, userID int64, offset, limit int) ([]*models.Settlement, error) {
	query := `
		SELECT s.id, s.uuid, s.group_id, s.from_user_id, s.to_user_id, s.amount, s.currency, s.description, s.created_at,
		       g.uuid as group_uuid, g.name as group_name,
		       fu.uuid as from_user_uuid, fu.name as from_user_name, fu.email as from_user_email,
		       tu.uuid as to_user_uuid, tu.name as to_user_name, tu.email as to_user_email
		FROM settlements s
		LEFT JOIN ` + "`groups`" + ` g ON s.group_id = g.id
		LEFT JOIN users fu ON s.from_user_id = fu.id
		LEFT JOIN users tu ON s.to_user_id = tu.id
		WHERE s.from_user_id = ? OR s.to_user_id = ?
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get user settlements", zap.Error(err), zap.Int64("userID", userID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var settlements []*models.Settlement
	for rows.Next() {
		settlement := &models.Settlement{}
		group := &models.Group{}
		fromUser := &models.User{}
		toUser := &models.User{}
		var groupUUID, groupName, fromUserUUID, fromUserName, fromUserEmail, toUserUUID, toUserName, toUserEmail sql.NullString

		err := rows.Scan(
			&settlement.ID, &settlement.UUID, &settlement.GroupID, &settlement.FromUserID, &settlement.ToUserID,
			&settlement.Amount, &settlement.Currency, &settlement.Description, &settlement.CreatedAt,
			&groupUUID, &groupName,
			&fromUserUUID, &fromUserName, &fromUserEmail,
			&toUserUUID, &toUserName, &toUserEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan user settlement row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if groupUUID.Valid {
			group.ID = settlement.GroupID
			group.UUID = groupUUID.String
			group.Name = groupName.String
			settlement.Group = group
		}

		if fromUserUUID.Valid {
			fromUser.ID = settlement.FromUserID
			fromUser.UUID = fromUserUUID.String
			fromUser.Name = fromUserName.String
			fromUser.Email = fromUserEmail.String
			settlement.FromUser = fromUser
		}

		if toUserUUID.Valid {
			toUser.ID = settlement.ToUserID
			toUser.UUID = toUserUUID.String
			toUser.Name = toUserName.String
			toUser.Email = toUserEmail.String
			settlement.ToUser = toUser
		}

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}
