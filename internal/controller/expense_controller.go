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

type ExpenseController struct {
	expenseService service.ExpenseService
	logger         *zap.Logger
}

// NewExpenseController creates a new expense controller
func NewExpenseController(expenseService service.ExpenseService, logger *zap.Logger) *ExpenseController {
	return &ExpenseController{
		expenseService: expenseService,
		logger:         logger,
	}
}

// CreateExpense handles expense creation with splits
// @Summary Create a new expense
// @Description Create a new expense with different split types (equal, exact, percentage)
// @Tags expenses
// @Accept json
// @Produce json
// @Param expense body models.CreateExpenseRequest true "Expense creation request"
// @Success 201 {object} response.APIResponse{data=models.Expense}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/expenses [post]
func (c *ExpenseController) CreateExpense(ctx *gin.Context) {
	var req models.CreateExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(ctx, "Invalid request body")
		return
	}

	expense, err := c.expenseService.CreateExpense(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create expense", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, expense)
}

// ListExpenses handles expense listing with filtering
// @Summary List expenses
// @Description Get paginated list of expenses with optional filtering
// @Tags expenses
// @Produce json
// @Param group_uuid query string false "Filter by group UUID"
// @Param user_uuid query string false "Filter by user UUID"
// @Param currency query string false "Filter by currency"
// @Param split_type query string false "Filter by split type"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=models.ExpenseListResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/expenses [get]
func (c *ExpenseController) ListExpenses(ctx *gin.Context) {
	// Parse filter parameters
	filter := &models.ExpenseFilter{
		GroupUUID: ctx.Query("group_uuid"),
		UserUUID:  ctx.Query("user_uuid"),
		Currency:  ctx.Query("currency"),
		Page:      1,
		Limit:     10,
	}

	// Parse split type
	if splitType := ctx.Query("split_type"); splitType != "" {
		filter.SplitType = models.SplitType(splitType)
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

	expenseResponse, err := c.expenseService.ListExpenses(ctx.Request.Context(), filter)
	if err != nil {
		c.logger.Error("Failed to list expenses", zap.Error(err))
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, expenseResponse)
}

// GetGroupExpenses handles retrieval of expenses for a specific group
// @Summary Get group expenses
// @Description Get paginated list of expenses for a specific group
// @Tags expenses
// @Produce json
// @Param uuid path string true "Group UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Expense,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/groups/{uuid}/expenses [get]
func (c *ExpenseController) GetGroupExpenses(ctx *gin.Context) {
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

	expenses, err := c.expenseService.GetGroupExpenses(ctx.Request.Context(), uuid, page, limit)
	if err != nil {
		c.logger.Error("Failed to get group expenses", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(expenses),
	}

	response.SuccessWithMeta(ctx, expenses, meta)
}

// GetUserExpenses handles retrieval of expenses for a specific user
// @Summary Get user expenses
// @Description Get paginated list of expenses paid by a specific user
// @Tags expenses
// @Produce json
// @Param uuid path string true "User UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse{data=[]models.Expense,meta=response.Meta}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/{uuid}/expenses [get]
func (c *ExpenseController) GetUserExpenses(ctx *gin.Context) {
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

	expenses, err := c.expenseService.GetUserExpenses(ctx.Request.Context(), uuid, page, limit)
	if err != nil {
		c.logger.Error("Failed to get user expenses", zap.Error(err), zap.String("uuid", uuid))
		response.Error(ctx, err)
		return
	}

	// Create meta information
	meta := &response.Meta{
		Page:  page,
		Limit: limit,
		Total: len(expenses),
	}

	response.SuccessWithMeta(ctx, expenses, meta)
}
