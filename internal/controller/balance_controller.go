package controller

import (
	"expense-split-tracker/internal/service"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceController struct {
	balanceService service.BalanceService
	logger         *zap.Logger
}

// NewBalanceController creates a new balance controller
func NewBalanceController(balanceService service.BalanceService, logger *zap.Logger) *BalanceController {
	return &BalanceController{
		balanceService: balanceService,
		logger:         logger,
	}
}

// GetBalanceSheet handles retrieval of group balance sheet
// @Summary Get group balance sheet
// @Description Get complete balance sheet for a group showing all user balances
// @Tags balances
// @Produce json
// @Param uuid path string true "Group UUID"
// @Success 200 {object} response.APIResponse{data=models.BalanceSheet}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/balance-sheet [get]
func (c *BalanceController) GetBalanceSheet(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	balanceSheet, err := c.balanceService.GetGroupBalanceSheet(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get balance sheet", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, balanceSheet)
}

// GetUserBalance handles retrieval of user balance in a group
// @Summary Get user balance in group
// @Description Get detailed balance information for a user in a specific group
// @Tags balances
// @Produce json
// @Param uuid path string true "Group UUID"
// @Param userUuid path string true "User UUID"
// @Success 200 {object} response.APIResponse{data=models.UserBalanceDetail}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/users/{userUuid}/balance [get]
func (c *BalanceController) GetUserBalance(ctx *gin.Context) {
	groupUuid := ctx.Param("uuid")
	userUuid := ctx.Param("userUuid")

	if groupUuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	if userUuid == "" {
		response.BadRequest(ctx, "User UUID is required")
		return
	}

	userBalance, err := c.balanceService.GetUserBalance(ctx.Request.Context(), groupUuid, userUuid)
	if err != nil {
		c.logger.Error("Failed to get user balance", zap.Error(err),
			zap.String("groupUuid", groupUuid), zap.String("userUuid", userUuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, userBalance)
}

// GetDebtRelationships handles retrieval of debt relationships in a group
// @Summary Get debt relationships
// @Description Get debt relationships between users in a group
// @Tags balances
// @Produce json
// @Param uuid path string true "Group UUID"
// @Success 200 {object} response.APIResponse{data=[]models.DebtRelationship}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/debt-relationships [get]
func (c *BalanceController) GetDebtRelationships(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		response.BadRequest(ctx, "Group UUID is required")
		return
	}

	relationships, err := c.balanceService.GetDebtRelationships(ctx.Request.Context(), uuid)
	if err != nil {
		c.logger.Error("Failed to get debt relationships", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, relationships)
}
