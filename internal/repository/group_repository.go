package repository

import (
	"context"
	"database/sql"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

type groupRepository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *database.DB, logger *zap.Logger) GroupRepository {
	return &groupRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new group
func (r *groupRepository) Create(ctx context.Context, tx *database.Tx, group *models.Group) error {
	query := `
		INSERT INTO ` + "`groups`" + ` (uuid, name, description, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, group.UUID, group.Name, group.Description, group.CreatedBy)
	} else {
		result, err = r.db.ExecContext(ctx, query, group.UUID, group.Name, group.Description, group.CreatedBy)
	}

	if err != nil {
		r.logger.Error("Failed to create group", zap.Error(err), zap.String("name", group.Name))
		return errors.NewDatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	group.ID = id
	r.logger.Info("Group created successfully", zap.Int64("id", group.ID), zap.String("name", group.Name))
	return nil
}

// GetByID retrieves a group by ID
func (r *groupRepository) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	query := `
		SELECT g.id, g.uuid, g.name, g.description, g.created_by, g.created_at, g.updated_at,
		       u.uuid as creator_uuid, u.name as creator_name, u.email as creator_email
		FROM ` + "`groups`" + ` g
		LEFT JOIN users u ON g.created_by = u.id
		WHERE g.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	group := &models.Group{}
	creator := &models.User{}
	var creatorUUID, creatorName, creatorEmail sql.NullString

	err := row.Scan(
		&group.ID, &group.UUID, &group.Name, &group.Description, &group.CreatedBy,
		&group.CreatedAt, &group.UpdatedAt,
		&creatorUUID, &creatorName, &creatorEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Group")
		}
		r.logger.Error("Failed to get group by ID", zap.Error(err), zap.Int64("id", id))
		return nil, errors.NewDatabaseError(err)
	}

	if creatorUUID.Valid {
		creator.ID = group.CreatedBy
		creator.UUID = creatorUUID.String
		creator.Name = creatorName.String
		creator.Email = creatorEmail.String
		group.Creator = creator
	}

	return group, nil
}

// GetByUUID retrieves a group by UUID
func (r *groupRepository) GetByUUID(ctx context.Context, uuid string) (*models.Group, error) {
	query := `
		SELECT g.id, g.uuid, g.name, g.description, g.created_by, g.created_at, g.updated_at,
		       u.uuid as creator_uuid, u.name as creator_name, u.email as creator_email
		FROM ` + "`groups`" + ` g
		LEFT JOIN users u ON g.created_by = u.id
		WHERE g.uuid = ?
	`

	row := r.db.QueryRowContext(ctx, query, uuid)

	group := &models.Group{}
	creator := &models.User{}
	var creatorUUID, creatorName, creatorEmail sql.NullString

	err := row.Scan(
		&group.ID, &group.UUID, &group.Name, &group.Description, &group.CreatedBy,
		&group.CreatedAt, &group.UpdatedAt,
		&creatorUUID, &creatorName, &creatorEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("Group")
		}
		r.logger.Error("Failed to get group by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, errors.NewDatabaseError(err)
	}

	if creatorUUID.Valid {
		creator.ID = group.CreatedBy
		creator.UUID = creatorUUID.String
		creator.Name = creatorName.String
		creator.Email = creatorEmail.String
		group.Creator = creator
	}

	return group, nil
}

// Update updates a group
func (r *groupRepository) Update(ctx context.Context, tx *database.Tx, group *models.Group) error {
	query := `
		UPDATE ` + "`groups`" + `
		SET name = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, group.Name, group.Description, group.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, group.Name, group.Description, group.ID)
	}

	if err != nil {
		r.logger.Error("Failed to update group", zap.Error(err), zap.Int64("id", group.ID))
		return errors.NewDatabaseError(err)
	}

	r.logger.Info("Group updated successfully", zap.Int64("id", group.ID))
	return nil
}

// Delete deletes a group
func (r *groupRepository) Delete(ctx context.Context, tx *database.Tx, id int64) error {
	query := `DELETE FROM ` + "`groups`" + ` WHERE id = ?`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, id)
	} else {
		result, err = r.db.ExecContext(ctx, query, id)
	}

	if err != nil {
		r.logger.Error("Failed to delete group", zap.Error(err), zap.Int64("id", id))
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Group")
	}

	r.logger.Info("Group deleted successfully", zap.Int64("id", id))
	return nil
}

// List retrieves a list of groups with pagination
func (r *groupRepository) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	query := `
		SELECT g.id, g.uuid, g.name, g.description, g.created_by, g.created_at, g.updated_at,
		       u.uuid as creator_uuid, u.name as creator_name, u.email as creator_email
		FROM ` + "`groups`" + ` g
		LEFT JOIN users u ON g.created_by = u.id
		ORDER BY g.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list groups", zap.Error(err))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		creator := &models.User{}
		var creatorUUID, creatorName, creatorEmail sql.NullString

		err := rows.Scan(
			&group.ID, &group.UUID, &group.Name, &group.Description, &group.CreatedBy,
			&group.CreatedAt, &group.UpdatedAt,
			&creatorUUID, &creatorName, &creatorEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan group row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if creatorUUID.Valid {
			creator.ID = group.CreatedBy
			creator.UUID = creatorUUID.String
			creator.Name = creatorName.String
			creator.Email = creatorEmail.String
			group.Creator = creator
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// GetUserGroups retrieves groups that a user is a member of
func (r *groupRepository) GetUserGroups(ctx context.Context, userID int64, offset, limit int) ([]*models.Group, error) {
	query := `
		SELECT g.id, g.uuid, g.name, g.description, g.created_by, g.created_at, g.updated_at,
		       u.uuid as creator_uuid, u.name as creator_name, u.email as creator_email
		FROM ` + "`groups`" + ` g
		LEFT JOIN users u ON g.created_by = u.id
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get user groups", zap.Error(err), zap.Int64("userID", userID))
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		creator := &models.User{}
		var creatorUUID, creatorName, creatorEmail sql.NullString

		err := rows.Scan(
			&group.ID, &group.UUID, &group.Name, &group.Description, &group.CreatedBy,
			&group.CreatedAt, &group.UpdatedAt,
			&creatorUUID, &creatorName, &creatorEmail,
		)
		if err != nil {
			r.logger.Error("Failed to scan user group row", zap.Error(err))
			return nil, errors.NewDatabaseError(err)
		}

		if creatorUUID.Valid {
			creator.ID = group.CreatedBy
			creator.UUID = creatorUUID.String
			creator.Name = creatorName.String
			creator.Email = creatorEmail.String
			group.Creator = creator
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// AddMember adds a user to a group
func (r *groupRepository) AddMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	query := `
		INSERT INTO group_members (group_id, user_id, joined_at)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE joined_at = joined_at
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, groupID, userID)
	} else {
		_, err = r.db.ExecContext(ctx, query, groupID, userID)
	}

	if err != nil {
		r.logger.Error("Failed to add member to group", zap.Error(err),
			zap.Int64("groupID", groupID), zap.Int64("userID", userID))
		return errors.NewDatabaseError(err)
	}

	r.logger.Info("Member added to group successfully",
		zap.Int64("groupID", groupID), zap.Int64("userID", userID))
	return nil
}

// RemoveMember removes a user from a group
func (r *groupRepository) RemoveMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	query := `DELETE FROM group_members WHERE group_id = ? AND user_id = ?`

	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.ExecContext(ctx, query, groupID, userID)
	} else {
		result, err = r.db.ExecContext(ctx, query, groupID, userID)
	}

	if err != nil {
		r.logger.Error("Failed to remove member from group", zap.Error(err),
			zap.Int64("groupID", groupID), zap.Int64("userID", userID))
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Group membership")
	}

	r.logger.Info("Member removed from group successfully",
		zap.Int64("groupID", groupID), zap.Int64("userID", userID))
	return nil
}

// GetMembers retrieves all members of a group
func (r *groupRepository) GetMembers(ctx context.Context, groupID int64) ([]*models.User, error) {
	query := `
		SELECT u.id, u.uuid, u.name, u.email, u.created_at, u.updated_at
		FROM users u
		INNER JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY gm.joined_at ASC
	`

	users := []*models.User{}
	err := r.db.SelectContext(ctx, &users, query, groupID)
	if err != nil {
		r.logger.Error("Failed to get group members", zap.Error(err), zap.Int64("groupID", groupID))
		return nil, errors.NewDatabaseError(err)
	}

	return users, nil
}

// IsMember checks if a user is a member of a group
func (r *groupRepository) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`

	var count int
	err := r.db.GetContext(ctx, &count, query, groupID, userID)
	if err != nil {
		r.logger.Error("Failed to check group membership", zap.Error(err),
			zap.Int64("groupID", groupID), zap.Int64("userID", userID))
		return false, errors.NewDatabaseError(err)
	}

	return count > 0, nil
}
