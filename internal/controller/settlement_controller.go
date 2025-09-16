package controller

import (
	"strconv"
	"time"

	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SettlementController struct {
	settlementService service.SettlementService
	logger            *zap.Logger
}

// NewSettlementController creates a new settlement controller
func NewSettlementController(settlementService service.SettlementService, logger *zap.Logger) *SettlementController {
	return &SettlementController{
		settlementService: settlementService,
		logger:            logger,
	}
}

// CreateSettlement handles settlement creation
// @Summary Create a new settlement
// @Description Create a new settlement (debt payment) between users
// @Tags settlements
// @Accept json
// @Produce json
// @Param settlement body models.CreateSettlementRequest true "Settlement creation request"
// @Success 201 {object} response.APIResponse{data=models.Settlement}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/settlements [post]
func (c *SettlementController) CreateSettlement(ctx *gin.Context) {
	var req models.CreateSettlementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(ctx, "Invalid request body")
		return
	}

	settlement, err := c.settlementService.CreateSettlement(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create settlement", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, settlement)
}

// GetSettlement handles settlement retrieval by UUID
// @Summary Get settlement by UUID
// @Description Get settlement details by UUID
// @Tags settlements
// @Produce json
// @Param uuid path string true "Settlement UUID"
// @Success 200 {object} response.APIResponse{data=models.Settlement}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/settlements/{uuid} [get]
func (c *SettlementController) GetSettlement(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Settlement UUID is required")
		return
	}

	settlement, err := c.settlementService.GetSettlementByUUID(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get settlement", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, settlement)
}

// ListSettlements handles settlement listing with filtering
// @Summary List settlements
// @Description Get paginated list of settlements with optional filtering
// @Tags settlements
// @Produce json
// @Param group_uuid query string false "Filter by group UUID"
// @Param user_uuid query string false "Filter by user UUID (either from or to)"
// @Param from_user_uuid query string false "Filter by from user UUID"
// @Param to_user_uuid query string false "Filter by to user UUID"
// @Param currency query string false "Filter by currency"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=models.SettlementListResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/settlements [get]
func (c *SettlementController) ListSettlements(ctx *gin.Context) {
	// Parse filter parameters
	filter := &models.SettlementFilter{
		GroupUUID:    ctx.Query("group_uuid"),
		UserUUID:     ctx.Query("user_uuid"),
		FromUserUUID: ctx.Query("from_user_uuid"),
		ToUserUUID:   ctx.Query("to_user_uuid"),
		Currency:     ctx.Query("currency"),
		Page:         1,
		Limit:        10,
	}

	// Parse dates
	if fromDateStr := ctx.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filter.FromDate = fromDate
		}
	}

	if toDateStr := ctx.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filter.ToDate = toDate
		}
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	settlementResponse, err := c.settlementService.ListSettlements(ctx.Request.Context(), filter)
	if err != nil {
		c.logger.Error("Failed to list settlements", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, settlementResponse)
}

// GetGroupSettlements handles retrieval of settlements for a specific group
// @Summary Get group settlements
// @Description Get paginated list of settlements for a specific group
// @Tags settlements
// @Produce json
// @Param uuid path string true "Group UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Settlement,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/settlements [get]
func (c *SettlementController) GetGroupSettlements(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
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

	settlements, err := c.settlementService.GetGroupSettlements(ctx.Request.Context(), uuid, page, limit)
	if err != nil {
		c.logger.Error("Failed to get group settlements", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(settlements),
	}

	response.SuccessWithMeta(ctx, settlements, meta)
}

// GetUserSettlements handles retrieval of settlements for a specific user
// @Summary Get user settlements
// @Description Get paginated list of settlements for a specific user (either as payer or receiver)
// @Tags settlements
// @Produce json
// @Param uuid path string true "User UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Settlement,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/{uuid}/settlements [get]
func (c *SettlementController) GetUserSettlements(ctx *gin.Context) {
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

	settlements, err := c.settlementService.GetUserSettlements(ctx.Request.Context(), uuid, page, limit)
	if err != nil {
		c.logger.Error("Failed to get user settlements", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(settlements),
	}

	response.SuccessWithMeta(ctx, settlements, meta)
}

// SimplifyDebts handles debt simplification for a group
// @Summary Simplify group debts
// @Description Get debt simplification suggestions for a group
// @Tags settlements
// @Produce json
// @Param uuid path string true "Group UUID"
// @Success 200 {object} response.APIResponse{data=models.DebtSimplification}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/simplify-debts [get]
func (c *SettlementController) SimplifyDebts(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	simplification, err := c.settlementService.SimplifyDebts(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to simplify debts", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, simplification)
}
