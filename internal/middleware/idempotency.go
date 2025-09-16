package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"expense-split-tracker/internal/config"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/utils"
	"expense-split-tracker/pkg/errors"
	"expense-split-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const IdempotencyKeyHeader = "Idempotency-Key"

type IdempotencyMiddleware struct {
	repo   repository.IdempotencyRepository
	config *config.Config
	logger *zap.Logger
}

// NewIdempotencyMiddleware creates a new idempotency middleware
func NewIdempotencyMiddleware(repo repository.IdempotencyRepository, config *config.Config, logger *zap.Logger) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		repo:   repo,
		config: config,
		logger: logger,
	}
}

// responseWriter wraps gin.ResponseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Handle processes idempotency for specific endpoints that need it
func (m *IdempotencyMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply idempotency to operations that actually need it
		if !m.shouldApplyIdempotency(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader(IdempotencyKeyHeader)
		if idempotencyKey == "" {
			// Idempotency key is required for financial operations
			response.Error(c, errors.NewValidationError("Idempotency-Key header is required for this operation"))
			c.Abort()
			return
		}

		// Validate idempotency key format (should be UUID)
		if !utils.IsValidUUID(idempotencyKey) {
			response.Error(c, errors.NewValidationError("Idempotency-Key must be a valid UUID"))
			c.Abort()
			return
		}

		// Read and hash request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			m.logger.Error("Failed to read request body", zap.Error(err))
			response.Error(c, errors.NewInternalError("Failed to process request"))
			c.Abort()
			return
		}

		// Restore request body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Create request hash including method, path, and body
		requestData := map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"body":   string(body),
		}

		requestHash, err := utils.HashRequest(requestData)
		if err != nil {
			m.logger.Error("Failed to hash request", zap.Error(err))
			response.Error(c, errors.NewInternalError("Failed to process request"))
			c.Abort()
			return
		}

		// Check if we've seen this idempotency key before
		existing, err := m.repo.GetByKey(c.Request.Context(), idempotencyKey)
		if err != nil {
			m.logger.Error("Failed to get idempotency record", zap.Error(err))
			response.Error(c, errors.NewInternalError("Failed to process request"))
			c.Abort()
			return
		}

		if existing != nil {
			// Check if the request hash matches
			if existing.RequestHash != requestHash {
				response.Error(c, errors.NewIdempotencyError("Idempotency key reused with different request"))
				c.Abort()
				return
			}

			// Return cached response
			c.Header("X-Idempotent-Replayed", "true")
			c.Data(existing.StatusCode, "application/json", existing.ResponseData)
			c.Abort()
			return
		}

		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
			status:         http.StatusOK,
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Store idempotency record after successful processing
		if !c.IsAborted() && writer.status < 500 {
			expiresAt := time.Now().Add(m.config.Features.IdempotencyTTL).Unix()

			err = m.repo.Create(
				c.Request.Context(),
				nil,
				idempotencyKey,
				requestHash,
				writer.body.Bytes(),
				writer.status,
				expiresAt,
			)

			if err != nil {
				m.logger.Error("Failed to store idempotency record",
					zap.Error(err),
					zap.String("key", idempotencyKey))
				// Don't fail the request, just log the error
			}
		}
	}
}

// shouldApplyIdempotency determines if idempotency should be applied to the request
// Only apply to financial operations that could cause duplicate charges/payments
func (m *IdempotencyMiddleware) shouldApplyIdempotency(method, path string) bool {
	method = strings.ToUpper(method)

	// Only apply to POST requests for financial operations
	if method != "POST" {
		return false
	}

	// Define endpoints that need idempotency (financial operations)
	idempotentEndpoints := []string{
		"/api/v1/expenses",    // Creating expenses - critical for financial accuracy
		"/api/v1/settlements", // Recording payments - critical for financial accuracy
		// "/api/v1/groups",      // Creating groups - temporarily disabled for testing
	}

	for _, endpoint := range idempotentEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

// CleanupExpiredKeys periodically cleans up expired idempotency keys
func (m *IdempotencyMiddleware) CleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := m.repo.DeleteExpired(context.TODO(), nil)
		if err != nil {
			m.logger.Error("Failed to cleanup expired idempotency keys", zap.Error(err))
		}
	}
}
