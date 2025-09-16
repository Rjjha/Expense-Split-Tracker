package routes

import (
	"net/http"

	"expense-split-tracker/internal/controller"
	"expense-split-tracker/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, services *service.Services, logger *zap.Logger) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "expense-split-tracker",
			"version": "1.0.0",
		})
	})

	// API version 1 routes
	v1 := router.Group("/api/v1")
	{
		setupUserRoutes(v1, services, logger)
		setupGroupRoutes(v1, services, logger)
		setupExpenseRoutes(v1, services, logger)
		setupSettlementRoutes(v1, services, logger)
		setupBalanceRoutes(v1, services, logger)
	}
}

// setupUserRoutes configures user-related routes
func setupUserRoutes(rg *gin.RouterGroup, services *service.Services, logger *zap.Logger) {
	userController := controller.NewUserController(services.User, logger)

	users := rg.Group("/users")
	{
		users.POST("", userController.CreateUser)
		users.GET("", userController.ListUsers)
		users.GET("/by-email", userController.GetUserByEmail)
		users.GET("/:uuid", userController.GetUser)
	}
}

// setupGroupRoutes configures group-related routes
func setupGroupRoutes(rg *gin.RouterGroup, services *service.Services, logger *zap.Logger) {
	groupController := controller.NewGroupController(services.Group, logger)

	groups := rg.Group("/groups")
	{
		groups.POST("", groupController.CreateGroup)
		groups.GET("", groupController.ListGroups)
		groups.GET("/:uuid", groupController.GetGroup)

		// Member management
		groups.POST("/:uuid/members", groupController.AddMember)
		groups.DELETE("/:uuid/members/:userUuid", groupController.RemoveMember)
		groups.GET("/:uuid/members", groupController.GetMembers)
	}

	// User's groups
	rg.GET("/users/:uuid/groups", groupController.GetUserGroups)
}

// setupExpenseRoutes configures expense-related routes
func setupExpenseRoutes(rg *gin.RouterGroup, services *service.Services, logger *zap.Logger) {
	expenseController := controller.NewExpenseController(services.Expense, logger)

	expenses := rg.Group("/expenses")
	{
		expenses.POST("", expenseController.CreateExpense)
		expenses.GET("", expenseController.ListExpenses)
	}

	// Group expenses
	rg.GET("/groups/:uuid/expenses", expenseController.GetGroupExpenses)
	// User expenses
	rg.GET("/users/:uuid/expenses", expenseController.GetUserExpenses)
}

// setupSettlementRoutes configures settlement-related routes
func setupSettlementRoutes(rg *gin.RouterGroup, services *service.Services, logger *zap.Logger) {
	settlementController := controller.NewSettlementController(services.Settlement, logger)

	settlements := rg.Group("/settlements")
	{
		settlements.POST("", settlementController.CreateSettlement)
		settlements.GET("", settlementController.ListSettlements)
		settlements.GET("/:uuid", settlementController.GetSettlement)
	}

	// Group settlements
	rg.GET("/groups/:uuid/settlements", settlementController.GetGroupSettlements)
	// User settlements
	rg.GET("/users/:uuid/settlements", settlementController.GetUserSettlements)
	// Debt simplification (read-only)
	rg.GET("/groups/:uuid/simplify-debts", settlementController.SimplifyDebts)
}

// setupBalanceRoutes configures balance-related routes
func setupBalanceRoutes(rg *gin.RouterGroup, services *service.Services, logger *zap.Logger) {
	balanceController := controller.NewBalanceController(services.Balance, logger)

	// Group balance sheet
	rg.GET("/groups/:uuid/balance-sheet", balanceController.GetBalanceSheet)
	// User balance in group (changed to avoid route conflict)
	rg.GET("/groups/:uuid/users/:userUuid/balance", balanceController.GetUserBalance)
	// Debt relationships
	rg.GET("/groups/:uuid/debt-relationships", balanceController.GetDebtRelationships)
}
