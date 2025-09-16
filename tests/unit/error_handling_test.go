package unit

import (
	"context"
	"testing"

	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestErrorHandling_InvalidUUIDs(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	es := service.NewExpenseService(nil, nil, nil, nil, nil, logger)
	s := service.NewSettlementService(nil, nil, nil, nil, nil, logger)
	bs := service.NewBalanceService(nil, nil, nil, nil, nil, logger)

	_, err := es.CreateExpense(ctx, &models.CreateExpenseRequest{GroupUUID: "bad", PaidByUUID: "bad", Amount: decimal.NewFromInt(1), Description: "d", SplitType: models.SplitTypeEqual, Splits: []models.CreateExpenseSplitRequest{{UserUUID: "bad"}}})
	assert.Error(t, err)

	_, err = s.CreateSettlement(ctx, &models.CreateSettlementRequest{GroupUUID: "bad", FromUserUUID: "bad", ToUserUUID: "bad", Amount: decimal.NewFromInt(1)})
	assert.Error(t, err)

	_, err = bs.GetGroupBalanceSheet(ctx, "bad")
	assert.Error(t, err)
}
