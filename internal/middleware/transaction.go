package middleware

import (
	"expense-split-tracker/internal/database"
	"expense-split-tracker/pkg/errors"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const TransactionKey = "db_transaction"

// TransactionMiddleware provides database transaction management
type TransactionMiddleware struct {
	db     *database.DB
	logger *zap.Logger
}

// NewTransactionMiddleware creates a new transaction middleware
func NewTransactionMiddleware(db *database.DB, logger *zap.Logger) *TransactionMiddleware {
	return &TransactionMiddleware{
		db:     db,
		logger: logger,
	}
}

// Handle wraps the request in a database transaction
func (m *TransactionMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply transactions to mutating operations
		if !m.shouldUseTransaction(c.Request.Method) {
			c.Next()
			return
		}

		tx, err := m.db.BeginTx()
		if err != nil {
			m.logger.Error("Failed to begin transaction", zap.Error(err))
			response.Error(c, errors.NewInternalError("Database transaction failed"))
			c.Abort()
			return
		}

		// Store transaction in context
		c.Set(TransactionKey, tx)

		// Defer transaction handling
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r) // re-throw panic after rollback
			} else if c.IsAborted() || len(c.Errors) > 0 {
				tx.Rollback()
			} else {
				if err := tx.Commit(); err != nil {
					m.logger.Error("Failed to commit transaction", zap.Error(err))
					response.Error(c, errors.NewInternalError("Database transaction commit failed"))
					c.Abort()
				}
			}
		}()

		c.Next()
	}
}

// GetTransaction retrieves the transaction from the gin context
func GetTransaction(c *gin.Context) *database.Tx {
	if tx, exists := c.Get(TransactionKey); exists {
		return tx.(*database.Tx)
	}
	return nil
}

// shouldUseTransaction determines if a transaction should be used for the request
func (m *TransactionMiddleware) shouldUseTransaction(method string) bool {
	mutatingMethods := map[string]bool{
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}
	return mutatingMethods[method]
}
