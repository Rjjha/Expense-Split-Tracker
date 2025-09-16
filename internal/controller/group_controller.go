package controller

import (
	"strconv"

	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GroupController struct {
	groupService service.GroupService
	logger       *zap.Logger
}

// NewGroupController creates a new group controller
func NewGroupController(groupService service.GroupService, logger *zap.Logger) *GroupController {
	return &GroupController{
		groupService: groupService,
		logger:       logger,
	}
}

// CreateGroup handles group creation
// @Summary Create a new group
// @Description Create a new group for expense tracking
// @Tags groups
// @Accept json
// @Produce json
// @Param group body models.CreateGroupRequest true "Group creation request"
// @Param creator_uuid query string true "UUID of the user creating the group"
// @Success 201 {object} response.APIResponse{data=models.Group}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups [post]
func (c *GroupController) CreateGroup(ctx *gin.Context) {
	var req models.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(ctx, "Invalid request body")
		return
	}

	creatorUUID := ctx.Query("creator_uuid")
	if creatorUUID == "" {
		response.BadRequest(ctx, "creator_uuid query parameter is required")
		return
	}

	group, err := c.groupService.CreateGroup(ctx.Request.Context(), &req, creatorUUID)
	if err != nil {
		c.logger.Error("Failed to create group", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, group)
}

// GetGroup handles group retrieval by UUID
// @Summary Get group by UUID
// @Description Get group details by UUID
// @Tags groups
// @Produce json
// @Param uuid path string true "Group UUID"
// @Success 200 {object} response.APIResponse{data=models.Group}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid} [get]
func (c *GroupController) GetGroup(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	group, err := c.groupService.GetGroupByUUID(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get group", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, group)
}

// ListGroups handles group listing with pagination
// @Summary List groups
// @Description Get paginated list of groups
// @Tags groups
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Group,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups [get]
func (c *GroupController) ListGroups(ctx *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	groups, err := c.groupService.ListGroups(ctx.Request.Context(), page, limit)
	if err != nil {
		c.logger.Error("Failed to list groups", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(groups),
	}

	response.SuccessWithMeta(ctx, groups, meta)
}

// GetUserGroups handles retrieval of groups for a specific user
// @Summary Get user's groups
// @Description Get paginated list of groups that a user is a member of
// @Tags groups
// @Produce json
// @Param uuid path string true "User UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Group,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/{uuid}/groups [get]
func (c *GroupController) GetUserGroups(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "User UUID is required")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	groups, err := c.groupService.GetUserGroups(ctx.Request.Context(), uuid, page, limit)
	if err != nil {
		c.logger.Error("Failed to get user groups", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(groups),
	}

	response.SuccessWithMeta(ctx, groups, meta)
}

// AddMember handles adding a member to a group
// @Summary Add member to group
// @Description Add a user as a member of a group
// @Tags groups
// @Accept json
// @Produce json
// @Param uuid path string true "Group UUID"
// @Param member body models.AddMemberRequest true "Add member request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/members [post]
func (c *GroupController) AddMember(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	var req models.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(ctx, "Invalid request body")
		return
	}

	err := c.groupService.AddMember(ctx.Request.Context(), uuid, &req)
	if err != nil {
		c.logger.Error("Failed to add member to group", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "Member added successfully"})
}

// RemoveMember handles removing a member from a group
// @Summary Remove member from group
// @Description Remove a user from a group
// @Tags groups
// @Produce json
// @Param uuid path string true "Group UUID"
// @Param userUuid path string true "User UUID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/members/{userUuid} [delete]
func (c *GroupController) RemoveMember(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	userUuid := ctx.Param("userUuid")

	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	if userUuid == "" {
		response.BadRequest(ctx, "User UUID is required")
		return
	}

	err := c.groupService.RemoveMember(ctx.Request.Context(), uuid, userUuid)
	if err != nil {
		c.logger.Error("Failed to remove member from group", zap.Error(err),
			zap.String("groupUuid", uuid), zap.String("userUuid", userUuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "Member removed successfully"})
}

// GetMembers handles retrieval of group members
// @Summary Get group members
// @Description Get all members of a group
// @Tags groups
// @Produce json
// @Param uuid path string true "Group UUID"
// @Success 200 {object} response.APIResponse{data=[]models.User}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/members [get]
func (c *GroupController) GetMembers(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	members, err := c.groupService.GetGroupMembers(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get group members", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, members)
}
