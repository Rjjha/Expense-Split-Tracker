package service

import (
	"context"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/repository"
	"expense-split-tracker/internal/utils"
	"expense-split-tracker/pkg/errors"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type settlementService struct {
	settlementRepo repository.SettlementRepository
	groupRepo      repository.GroupRepository
	userRepo       repository.UserRepository
	balanceRepo    repository.BalanceRepository
	db             DBTransactor
	logger         *zap.Logger
}

// NewSettlementService creates a new settlement service
func NewSettlementService(
	settlementRepo repository.SettlementRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	balanceRepo repository.BalanceRepository,
	db DBTransactor,
	logger *zap.Logger,
) SettlementService {
	return &settlementService{
		settlementRepo: settlementRepo,
		groupRepo:      groupRepo,
		userRepo:       userRepo,
		balanceRepo:    balanceRepo,
		db:             db,
		logger:         logger,
	}
}

// CreateSettlement creates a new settlement (debt payment)
func (s *settlementService) CreateSettlement(ctx context.Context, req *models.CreateSettlementRequest) (*models.Settlement, error) {
	// Validate input
	if err := utils.ValidateAmount(req.Amount); err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	if err := utils.ValidateCurrency(currency); err != nil {
		return nil, err
	}

	if !utils.IsValidUUID(req.GroupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", req.GroupUUID)
	}

	if !utils.IsValidUUID(req.FromUserUUID) {
		return nil, errors.NewInvalidValueError("from_user_uuid", req.FromUserUUID)
	}

	if !utils.IsValidUUID(req.ToUserUUID) {
		return nil, errors.NewInvalidValueError("to_user_uuid", req.ToUserUUID)
	}

	if req.FromUserUUID == req.ToUserUUID {
		return nil, errors.NewValidationError("From user and to user cannot be the same")
	}

	// Get group and validate
	group, err := s.groupRepo.GetByUUID(ctx, req.GroupUUID)
	if err != nil {
		return nil, err
	}

	// Get users and validate
	fromUser, err := s.userRepo.GetByUUID(ctx, req.FromUserUUID)
	if err != nil {
		return nil, err
	}

	toUser, err := s.userRepo.GetByUUID(ctx, req.ToUserUUID)
	if err != nil {
		return nil, err
	}

	// Check if both users are members of the group
	isFromMember, err := s.groupRepo.IsMember(ctx, group.ID, fromUser.ID)
	if err != nil {
		return nil, err
	}
	if !isFromMember {
		return nil, errors.NewValidationError("From user must be a member of the group")
	}

	isToMember, err := s.groupRepo.IsMember(ctx, group.ID, toUser.ID)
	if err != nil {
		return nil, err
	}
	if !isToMember {
		return nil, errors.NewValidationError("To user must be a member of the group")
	}

	// Get current balances to validate settlement amount
	fromBalance, err := s.balanceRepo.GetByGroupAndUser(ctx, group.ID, fromUser.ID, currency)
	if err != nil {
		return nil, err
	}

	// Validate settlement amount (user cannot pay more than they owe)
	if req.Amount.GreaterThan(fromBalance.Balance) {
		return nil, errors.NewInsufficientFundError(
			fromBalance.Balance.String(),
			req.Amount.String(),
		)
	}

	// Create settlement with transaction
	settlement := &models.Settlement{
		UUID:        utils.GenerateUUID(),
		GroupID:     group.ID,
		FromUserID:  fromUser.ID,
		ToUserID:    toUser.ID,
		Amount:      req.Amount,
		Currency:    currency,
		Description: req.Description,
	}

	err = s.db.WithTransaction(func(tx *database.Tx) error {
		// Create settlement
		if err := s.settlementRepo.Create(ctx, tx, settlement); err != nil {
			return err
		}

		// Update balances
		return s.updateBalancesAfterSettlement(ctx, tx, settlement)
	})

	if err != nil {
		s.logger.Error("Failed to create settlement", zap.Error(err))
		return nil, err
	}

	// Get complete settlement with relationships
	settlement, err = s.settlementRepo.GetByUUID(ctx, settlement.UUID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Settlement created successfully", zap.String("uuid", settlement.UUID))
	return settlement, nil
}

// updateBalancesAfterSettlement updates user balances after creating a settlement
func (s *settlementService) updateBalancesAfterSettlement(ctx context.Context, tx *database.Tx, settlement *models.Settlement) error {
	// Reduce debt for the payer (fromUser owes less)
	err := s.balanceRepo.UpdateBalance(ctx, tx, settlement.GroupID, settlement.FromUserID, settlement.Amount.Neg(), settlement.Currency)
	if err != nil {
		return err
	}

	// Reduce credit for the receiver (toUser is owed less)
	err = s.balanceRepo.UpdateBalance(ctx, tx, settlement.GroupID, settlement.ToUserID, settlement.Amount, settlement.Currency)
	if err != nil {
		return err
	}

	return nil
}

// GetSettlementByUUID retrieves a settlement by UUID
func (s *settlementService) GetSettlementByUUID(ctx context.Context, uuid string) (*models.Settlement, error) {
	if !utils.IsValidUUID(uuid) {
		return nil, errors.NewInvalidValueError("uuid", uuid)
	}

	settlement, err := s.settlementRepo.GetByUUID(ctx, uuid)
	if err != nil {
		s.logger.Error("Failed to get settlement by UUID", zap.Error(err), zap.String("uuid", uuid))
		return nil, err
	}

	return settlement, nil
}

// ListSettlements retrieves settlements with filtering
func (s *settlementService) ListSettlements(ctx context.Context, filter *models.SettlementFilter) (*models.SettlementListResponse, error) {
	settlements, total, err := s.settlementRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list settlements", zap.Error(err))
		return nil, err
	}

	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return &models.SettlementListResponse{
		Settlements: settlements,
		TotalCount:  total,
		Page:        page,
		Limit:       limit,
	}, nil
}

// GetGroupSettlements retrieves settlements for a specific group
func (s *settlementService) GetGroupSettlements(ctx context.Context, groupUUID string, page, limit int) ([]*models.Settlement, error) {
	if !utils.IsValidUUID(groupUUID) {
		return nil, errors.NewInvalidValueError("group_uuid", groupUUID)
	}

	group, err := s.groupRepo.GetByUUID(ctx, groupUUID)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	settlements, err := s.settlementRepo.GetGroupSettlements(ctx, group.ID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get group settlements", zap.Error(err), zap.String("groupUUID", groupUUID))
		return nil, err
	}

	return settlements, nil
}

// GetUserSettlements retrieves settlements for a specific user
func (s *settlementService) GetUserSettlements(ctx context.Context, userUUID string, page, limit int) ([]*models.Settlement, error) {
	if !utils.IsValidUUID(userUUID) {
		return nil, errors.NewInvalidValueError("user_uuid", userUUID)
	}

	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	settlements, err := s.settlementRepo.GetUserSettlements(ctx, user.ID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get user settlements", zap.Error(err), zap.String("userUUID", userUUID))
		return nil, err
	}

	return settlements, nil
}

// SimplifyDebts calculates debt simplification suggestions for a group
func (s *settlementService) SimplifyDebts(ctx context.Context, groupUUID string) (*models.DebtSimplification, error) {
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

	// Separate creditors (negative balance - they are owed money) and debtors (positive balance - they owe money)
	var creditors, debtors []*models.Balance
	for _, balance := range balances {
		if balance.Balance.GreaterThan(decimal.Zero) {
			debtors = append(debtors, balance)
		} else if balance.Balance.LessThan(decimal.Zero) {
			// Convert to positive for easier calculation
			balance.Balance = balance.Balance.Abs()
			creditors = append(creditors, balance)
		}
	}

	// Calculate minimum number of transactions needed
	originalTransactions := len(debtors) * len(creditors) // Worst case: everyone owes everyone
	if originalTransactions == 0 {
		originalTransactions = 1 // At least 1 to avoid division by zero
	}

	// Generate settlement suggestions using greedy algorithm
	suggestions := s.generateSettlementSuggestions(creditors, debtors, currency)

	simplifiedTransactions := len(suggestions)
	savings := originalTransactions - simplifiedTransactions
	if savings < 0 {
		savings = 0
	}

	return &models.DebtSimplification{
		OriginalTransactions:   originalTransactions,
		SimplifiedTransactions: simplifiedTransactions,
		Savings:                savings,
		Suggestions:            suggestions,
	}, nil
}

// generateSettlementSuggestions generates optimal settlement suggestions
func (s *settlementService) generateSettlementSuggestions(creditors, debtors []*models.Balance, currency string) []*models.SettlementSuggestion {
	var suggestions []*models.SettlementSuggestion

	// Create working copies
	creditorsWork := make([]*models.Balance, len(creditors))
	debtorsWork := make([]*models.Balance, len(debtors))

	for i, c := range creditors {
		creditorsWork[i] = &models.Balance{
			User:    c.User,
			Balance: c.Balance,
		}
	}

	for i, d := range debtors {
		debtorsWork[i] = &models.Balance{
			User:    d.User,
			Balance: d.Balance,
		}
	}

	// Greedy algorithm: always match the largest debtor with the largest creditor
	for len(creditorsWork) > 0 && len(debtorsWork) > 0 {
		// Find largest creditor and debtor
		maxCreditorIdx := 0
		maxDebtorIdx := 0

		for i, c := range creditorsWork {
			if c.Balance.GreaterThan(creditorsWork[maxCreditorIdx].Balance) {
				maxCreditorIdx = i
			}
		}

		for i, d := range debtorsWork {
			if d.Balance.GreaterThan(debtorsWork[maxDebtorIdx].Balance) {
				maxDebtorIdx = i
			}
		}

		creditor := creditorsWork[maxCreditorIdx]
		debtor := debtorsWork[maxDebtorIdx]

		// Calculate settlement amount (minimum of what creditor is owed and what debtor owes)
		settlementAmount := creditor.Balance
		if debtor.Balance.LessThan(settlementAmount) {
			settlementAmount = debtor.Balance
		}

		// Create suggestion
		suggestions = append(suggestions, &models.SettlementSuggestion{
			FromUser: debtor.User,
			ToUser:   creditor.User,
			Amount:   settlementAmount,
			Currency: currency,
		})

		// Update balances
		creditor.Balance = creditor.Balance.Sub(settlementAmount)
		debtor.Balance = debtor.Balance.Sub(settlementAmount)

		// Remove users with zero balance
		if creditor.Balance.Equal(decimal.Zero) {
			creditorsWork = append(creditorsWork[:maxCreditorIdx], creditorsWork[maxCreditorIdx+1:]...)
		}

		if debtor.Balance.Equal(decimal.Zero) {
			debtorsWork = append(debtorsWork[:maxDebtorIdx], debtorsWork[maxDebtorIdx+1:]...)
		}
	}

	return suggestions
}
