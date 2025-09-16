package service

import (
	"context"
	"time"

	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/utils"
	"expense-split-tracker/pkg/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type balanceService struct {
	balanceRepo    repository.BalanceRepository
	groupRepo      repository.GroupRepository
	userRepo       repository.UserRepository
	settlementRepo repository.SettlementRepository
	db             DBTransactor
	logger         *zap.Logger
}

// NewBalanceService creates a new balance service
func NewBalanceService(
	balanceRepo repository.BalanceRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	settlementRepo repository.SettlementRepository,
	db DBTransactor,
	logger *zap.Logger,
) BalanceService {
	return &balanceService{
		balanceRepo:    balanceRepo,
		groupRepo:      groupRepo,
		userRepo:       userRepo,
		settlementRepo: settlementRepo,
		db:             db,
		logger:         logger,
	}
}

// GetGroupBalanceSheet retrieves the complete balance sheet for a group
func (s *balanceService) GetGroupBalanceSheet(ctx context.Context, groupUUID string) (*models.BalanceSheet, error) {
	if !utils.IsValidUUID(groupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	// Get all balances for the group (assuming USD for now)
	currency := "USD"
	balances, err := s.balanceRepo.GetGroupBalances(ctx, group.ID, currency)
	if err != nil {
		return nil, err
	}

	// Convert to UserBalance format
	var userBalances []*models.UserBalance
	totalPositive := decimal.Zero
	totalNegative := decimal.Zero

	for _, balance := range balances {
		userBalance := &models.UserBalance{
			UserID:   balance.UserID,
			GroupID:  balance.GroupID,
			User:     balance.User,
			Balance:  balance.Balance,
			Currency: balance.Currency,
		}
		userBalances = append(userBalances, userBalance)

		if balance.Balance.GreaterThan(decimal.Zero) {
			totalPositive = totalPositive.Add(balance.Balance)
		} else if balance.Balance.LessThan(decimal.Zero) {
			totalNegative = totalNegative.Add(balance.Balance.Abs())
		}
	}

	// Create summary
	summary := &models.BalanceSummary{
		TotalPositive: totalPositive,
		TotalNegative: totalNegative,
		NetBalance:    totalPositive.Sub(totalNegative), // Should be close to zero in a balanced system
		UserCount:     len(userBalances),
	}

	balanceSheet := &models.BalanceSheet{
		Group:     group,
		Balances:  userBalances,
		Summary:   summary,
		Currency:  currency,
		UpdatedAt: time.Now(),
	}

	return balanceSheet, nil
}

// GetUserBalance retrieves detailed balance information for a user in a group
func (s *balanceService) GetUserBalance(ctx context.Context, groupUUID, userUUID string) (*models.UserBalanceDetail, error) {
	if !utils.IsValidUUID(groupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	if !utils.IsValidUUID(userUUID) {
		return nil, errors.NewInvalidValueError("user_uuid", userUUID)
	}

	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the group
	isMember, err := s.groupRepo.IsMember(ctx, group.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.NewValidationError("User is not a member of this group")
	}

	// Get current balance
	currency := "USD"
	balance, err := s.balanceRepo.GetByGroupAndUser(ctx, group.ID, user.ID, currency)
	if err != nil {
		return nil, err
	}

	// Get recent settlements for this user
	settlementFilter := &models.SettlementFilter{
		GroupUUID: groupUUID,
		UserUUID:  userUUID,
		Page:      1,
		Limit:     5, // Last 5 settlements
	}

	settlements, _, err := s.settlementRepo.List(ctx, settlementFilter)
	if err != nil {
		return nil, err
	}

	// Calculate breakdown (this is simplified - in a real system you'd query expenses and settlements)
	breakdown := &models.BalanceBreakdown{
		TotalPaid:    decimal.Zero, // TODO: Calculate from expenses where user is payer
		TotalOwed:    balance.Balance.Abs(),
		TotalSettled: decimal.Zero, // TODO: Calculate from settlements
		ExpenseCount: 0,            // TODO: Count expenses involving this user
		PaymentCount: len(settlements),
	}

	userBalanceDetail := &models.UserBalanceDetail{
		User:         user,
		Balance:      balance.Balance,
		Currency:     currency,
		Breakdown:    breakdown,
		Settlements:  settlements,
		LastActivity: balance.LastUpdated,
	}

	return userBalanceDetail, nil
}

// GetDebtRelationships retrieves debt relationships between users in a group
func (s *balanceService) GetDebtRelationships(ctx context.Context, groupUUID string) ([]*models.DebtRelationship, error) {
	if !utils.IsValidUUID(groupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	// Get all balances for the group
	currency := "USD"
	balances, err := s.balanceRepo.GetGroupBalances(ctx, group.ID, currency)
	if err != nil {
		return nil, err
	}

	// Separate creditors and debtors
	var creditors, debtors []*models.Balance
	for _, balance := range balances {
		if balance.Balance.GreaterThan(decimal.Zero) {
			debtors = append(debtors, balance)
		} else if balance.Balance.LessThan(decimal.Zero) {
			creditors = append(creditors, balance)
		}
	}

	// Create debt relationships
	var relationships []*models.DebtRelationship

	// Simple approach: each debtor owes proportionally to each creditor
	for _, debtor := range debtors {
		totalCredit := decimal.Zero
		for _, creditor := range creditors {
			totalCredit = totalCredit.Add(creditor.Balance.Abs())
		}

		if totalCredit.GreaterThan(decimal.Zero) {
			for _, creditor := range creditors {
				// Calculate proportional debt
				proportion := creditor.Balance.Abs().Div(totalCredit)
				debtAmount := debtor.Balance.Mul(proportion).Round(2)

				if debtAmount.GreaterThan(decimal.NewFromFloat(0.01)) { // Only include debts > 1 cent
					relationships = append(relationships, &models.DebtRelationship{
						Creditor: creditor.User,
						Debtor:   debtor.User,
						Amount:   debtAmount,
						Currency: currency,
					})
				}
			}
		}
	}

	return relationships, nil
}
