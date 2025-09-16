package controller

import (
	"strconv"

	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserController creates a new user controller
func NewUserController(userService service.UserService, logger *zap.Logger) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser handles user creation
// @Summary Create a new user
// @Description Create a new user with name and email
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User creation request"
// @Success 201 {object} response.APIResponse{data=models.User}
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(ctx, "Invalid request body")
		return
	}

	user, err := c.userService.CreateUser(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create user", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, user)
}

// GetUser handles user retrieval by UUID
// @Summary Get user by UUID
// @Description Get user details by UUID
// @Tags users
// @Produce json
// @Param uuid path string true "User UUID"
// @Success 200 {object} response.APIResponse{data=models.User}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/{uuid} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "User UUID is required")
		return
	}

	user, err := c.userService.GetUserByUUID(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get user", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, user)
}

// ListUsers handles user listing with pagination
// @Summary List users
// @Description Get paginated list of users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.User,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
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

	users, err := c.userService.ListUsers(ctx.Request.Context(), page, limit)
	if err != nil {
		c.logger.Error("Failed to list users", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(users), // This would ideally come from the service with a count query
	}

	response.SuccessWithMeta(ctx, users, meta)
}

// GetUserByEmail handles user retrieval by email
// @Summary Get user by email
// @Description Get user details by email address
// @Tags users
// @Produce json
// @Param email query string true "User email"
// @Success 200 {object} response.APIResponse{data=models.User}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/by-email [get]
func (c *UserController) GetUserByEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		response.BadRequest(ctx, "Email parameter is required")
		return
	}

	user, err := c.userService.GetUserByEmail(ctx.Request.Context(), email)
	if err != nil {
		c.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, user)
}
