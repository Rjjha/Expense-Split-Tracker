package service

import (
	"context"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/utils"
	"expense-split-tracker/pkg/errors"

	"go.uber.org/zap"
)

type groupService struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
	db        DBTransactor
	logger    *zap.Logger
}

// NewGroupService creates a new group service
func NewGroupService(groupRepo repository.GroupRepository, userRepo repository.UserRepository, db DBTransactor, logger *zap.Logger) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
		db:        db,
		logger:    logger,
	}
}

// CreateGroup creates a new group
func (s *groupService) CreateGroup(ctx context.Context, req *models.CreateGroupRequest, creatorUUID string) (*models.Group, error) {
	// Validate input
	if err := utils.ValidateName(req.Name); err != nil {
		return nil, err
	}

	if !utils.IsValidUUID(creatorUUID) {
		return nil, errors.NewInvalidValueError("creator_uuid", creatorUUID)
	}

	// Get creator user
	creator, err := s.userRepo.GetByUUID(ctx, creatorUUID)
	if err != nil {
		return nil, err
	}

	// Create group with transaction
	group := &models.Group{
		UUID:        utils.GenerateUUID(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   creator.ID,
	}

	err = s.db.WithTransaction(func(tx *database.Tx) error {
		// Create group
		if err := s.groupRepo.Create(ctx, tx, group); err != nil {
			return err
		}

		// Add creator as first member
		if err := s.groupRepo.AddMember(ctx, tx, group.ID, creator.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create group", zap.Error(err), zap.String("name", req.Name))
		return nil, err
	}

	group.Creator = creator
	s.logger.Info("Group created successfully", zap.String("uuid", group.UUID), zap.String("name", group.Name))
	return group, nil
}

// GetGroupByUUID retrieves a group by UUID
func (s *groupService) GetGroupByUUID(ctx context.Context, uuid string) (*models.Group, error) {
	if !utils.IsValidUUID(uuid) {
		return nil, errors.NewInvalidValueError("uuid", uuid)
	}

	group, err := s.groupRepo.GetByUUID(ctx, uuid)
	if err != nil {
		s.logger.Error("Failed to get group by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, err
	}

	// Get members
	members, err := s.groupRepo.GetMembers(ctx, group.ID)
	if err != nil {
		s.logger.Error("Failed to get group members", zap.Error(err), zap.Int64("groupID", group.ID))
		return nil, err
	}

	group.Members = members
	return group, nil
}

// ListGroups retrieves a paginated list of groups
func (s *groupService) ListGroups(ctx context.Context, page, limit int) ([]*models.Group, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	groups, err := s.groupRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list groups", zap.Error(err))
		return nil, err
	}

	return groups, nil
}

// GetUserGroups retrieves groups that a user is a member of
func (s *groupService) GetUserGroups(ctx context.Context, userUUID string, page, limit int) ([]*models.Group, error) {
	if !utils.IsValidUUID(userUUID) {
		return nil, errors.NewInvalidValueError("user_uuid", userUUID)
	}

	// Get user
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	groups, err := s.groupRepo.GetUserGroups(ctx, user.ID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get user groups", zap.Error(err), zap.String("userUUID", userUUID))
		return nil, err
	}

	return groups, nil
}

// AddMember adds a user to a group
func (s *groupService) AddMember(ctx context.Context, groupUUID string, req *models.AddMemberRequest) error {
	if !utils.IsValidUUID(groupUUID) {
		return errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	if !utils.IsValidUUID(req.UserUUID) {
		return errors.NewInvalidValueError("user_uuid", req.UserUUID)
	}

	// Get group
	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByUUID(ctx, req.UserUUID)
	if err != nil {
		return err
	}

	// Check if user is already a member
	isMember, err := s.groupRepo.IsMember(ctx, group.ID, user.ID)
	if err != nil {
		return err
	}

	if isMember {
		return errors.NewAlreadyExistsError("User is already a member of this group")
	}

	// Add member with transaction
	err = s.db.WithTransaction(func(tx *database.Tx) error {
		return s.groupRepo.AddMember(ctx, tx, group.ID, user.ID)
	})

	if err != nil {
		s.logger.Error("Failed to add member to group", zap.Error(err),
			zap.String("groupUUID", groupUUID), zap.String("userUUID", req.UserUUID))
		return err
	}

	s.logger.Info("Member added to group successfully",
		zap.String("groupUUID", groupUUID), zap.String("userUUID", req.UserUUID))
	return nil
}

// RemoveMember removes a user from a group
func (s *groupService) RemoveMember(ctx context.Context, groupUUID, userUUID string) error {
	if !utils.IsValidUUID(groupUUID) {
		return errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	if !utils.IsValidUUID(userUUID) {
		return errors.NewInvalidValueError("user_uuid", userUUID)
	}

	// Get group
	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return err
	}

	// Remove member with transaction
	err = s.db.WithTransaction(func(tx *database.Tx) error {
		return s.groupRepo.RemoveMember(ctx, tx, group.ID, user.ID)
	})

	if err != nil {
		s.logger.Error("Failed to remove member from group", zap.Error(err),
			zap.String("groupUUID", groupUUID), zap.String("userUUID", userUUID))
		return err
	}

	s.logger.Info("Member removed from group successfully",
		zap.String("groupUUID", groupUUID), zap.String("userUUID", userUUID))
	return nil
}

// GetGroupMembers retrieves all members of a group
func (s *groupService) GetGroupMembers(ctx context.Context, groupUUID string) ([]*models.User, error) {
	if !utils.IsValidUUID(groupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	// Get group
	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	members, err := s.groupRepo.GetMembers(ctx, group.ID)
	if err != nil {
		s.logger.Error("Failed to get group members", zap.Error(err), zap.String("groupUUID", groupUUID))
		return nil, err
	}

	return members, nil
}
