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

type expenseService struct {
	expenseRepo repository.ExpenseRepository
	groupRepo   repository.GroupRepository
	userRepo    repository.UserRepository
	balanceRepo repository.BalanceRepository
	db          DBTransactor
	logger      *zap.Logger
}

// NewExpenseService creates a new expense service
func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	balanceRepo repository.BalanceRepository,
	db DBTransactor,
	logger *zap.Logger,
) ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
		db:          db,
		logger:      logger,
	}
}

// CreateExpense creates a new expense with splits
func (s *expenseService) CreateExpense(ctx context.Context, req *models.CreateExpenseRequest) (*models.Expense, error) {
	// Validate input
	if err := utils.ValidateAmount(req.Amount); err != nil {
		return nil, err
	}

	if err := utils.ValidateDescription(req.Description); err != nil {
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

	if !utils.IsValidUUID(req.PaidByUUID) {
		return nil, errors.NewInvalidValueError("paid_by_uuid", req.PaidByUUID)
	}

	// Get group and validate
	group, err := s.groupRepo.GetByUUID(ctx, req.GroupUUID)
	if err != nil {
		return nil, err
	}

	// Get payer and validate
	payer, err := s.userRepo.GetByUUID(ctx, req.PaidByUUID)
	if err != nil {
		return nil, err
	}

	// Check if payer is a member of the group
	isMember, err := s.groupRepo.IsMember(ctx, group.ID, payer.ID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.NewValidationError("Payer must be a member of the group")
	}

	// Validate splits based on split type
	splits, err := s.validateAndCalculateSplits(ctx, req, group.ID)
	if err != nil {
		return nil, err
	}

	// Create expense with transaction
	expense := &models.Expense{
		UUID:        utils.GenerateUUID(),
		GroupID:     group.ID,
		PaidBy:      payer.ID,
		Amount:      req.Amount,
		Currency:    currency,
		Description: req.Description,
		SplitType:   req.SplitType,
	}

	err = s.db.WithTransaction(func(tx *database.Tx) error {
		// Create expense
		if err := s.expenseRepo.Create(ctx, tx, expense); err != nil {
			return err
		}

		// Create splits
		for _, split := range splits {
			split.ExpenseID = expense.ID
			if err := s.expenseRepo.CreateSplit(ctx, tx, split); err != nil {
				return err
			}
		}

		// Update balances
		return s.updateBalancesAfterExpense(ctx, tx, expense, splits)
	})

	if err != nil {
		s.logger.Error("Failed to create expense", zap.Error(err), zap.String("description", req.Description))
		return nil, err
	}

	// Get splits for response
	expense.Splits, err = s.expenseRepo.GetExpenseSplits(ctx, expense.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Expense created successfully", zap.String("uuid", expense.UUID), zap.String("description", expense.Description))
	return expense, nil
}

// validateAndCalculateSplits validates and calculates splits based on split type
func (s *expenseService) validateAndCalculateSplits(ctx context.Context, req *models.CreateExpenseRequest, groupID int64) ([]*models.ExpenseSplit, error) {
	if len(req.Splits) == 0 {
		return nil, errors.NewValidationError("At least one split is required")
	}

	var splits []*models.ExpenseSplit

	switch req.SplitType {
	case models.SplitTypeEqual:
		return s.calculateEqualSplits(ctx, req, groupID)

	case models.SplitTypeExact:
		return s.calculateExactSplits(ctx, req, groupID)

	case models.SplitTypePercentage:
		return s.calculatePercentageSplits(ctx, req, groupID)

	default:
		return nil, errors.NewInvalidValueError("split_type", string(req.SplitType))
	}

	return splits, nil
}

// calculateEqualSplits calculates equal splits among users
func (s *expenseService) calculateEqualSplits(ctx context.Context, req *models.CreateExpenseRequest, groupID int64) ([]*models.ExpenseSplit, error) {
	var splits []*models.ExpenseSplit
	splitCount := decimal.NewFromInt(int64(len(req.Splits)))
	amountPerUser := req.Amount.Div(splitCount).Round(2)

	// Handle rounding by giving remainder to first user
	totalAssigned := decimal.Zero

	for i, splitReq := range req.Splits {
		if !utils.IsValidUUID(splitReq.UserUUID) {
			return nil, errors.NewInvalidValueError("user_uuid", splitReq.UserUUID)
		}

		user, err := s.userRepo.GetByUUID(ctx, splitReq.UserUUID)
		if err != nil {
			return nil, err
		}

		// Check if user is a member of the group
		isMember, err := s.groupRepo.IsMember(ctx, groupID, user.ID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.NewValidationError("All users in split must be members of the group")
		}

		amount := amountPerUser

		// For the last user, assign remaining amount to handle rounding
		if i == len(req.Splits)-1 {
			amount = req.Amount.Sub(totalAssigned)
		}

		splits = append(splits, &models.ExpenseSplit{
			UserID: user.ID,
			Amount: amount,
			User:   user,
		})

		totalAssigned = totalAssigned.Add(amount)
	}

	return splits, nil
}

// calculateExactSplits calculates exact amount splits
func (s *expenseService) calculateExactSplits(ctx context.Context, req *models.CreateExpenseRequest, groupID int64) ([]*models.ExpenseSplit, error) {
	var splits []*models.ExpenseSplit
	totalSplitAmount := decimal.Zero

	for _, splitReq := range req.Splits {
		if !utils.IsValidUUID(splitReq.UserUUID) {
			return nil, errors.NewInvalidValueError("user_uuid", splitReq.UserUUID)
		}

		if splitReq.Amount.LessThanOrEqual(decimal.Zero) {
			return nil, errors.NewValidationError("Split amounts must be greater than zero")
		}

		user, err := s.userRepo.GetByUUID(ctx, splitReq.UserUUID)
		if err != nil {
			return nil, err
		}

		// Check if user is a member of the group
		isMember, err := s.groupRepo.IsMember(ctx, groupID, user.ID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.NewValidationError("All users in split must be members of the group")
		}

		splits = append(splits, &models.ExpenseSplit{
			UserID: user.ID,
			Amount: splitReq.Amount,
			User:   user,
		})

		totalSplitAmount = totalSplitAmount.Add(splitReq.Amount)
	}

	// Validate that split amounts equal total expense amount
	if !totalSplitAmount.Equal(req.Amount) {
		return nil, errors.NewInvalidSplitError("Sum of split amounts must equal total expense amount")
	}

	return splits, nil
}

// calculatePercentageSplits calculates percentage-based splits
func (s *expenseService) calculatePercentageSplits(ctx context.Context, req *models.CreateExpenseRequest, groupID int64) ([]*models.ExpenseSplit, error) {
	var splits []*models.ExpenseSplit
	totalPercentage := decimal.Zero

	for _, splitReq := range req.Splits {
		if !utils.IsValidUUID(splitReq.UserUUID) {
			return nil, errors.NewInvalidValueError("user_uuid", splitReq.UserUUID)
		}

		if err := utils.ValidatePercentage(splitReq.Percentage); err != nil {
			return nil, err
		}

		user, err := s.userRepo.GetByUUID(ctx, splitReq.UserUUID)
		if err != nil {
			return nil, err
		}

		// Check if user is a member of the group
		isMember, err := s.groupRepo.IsMember(ctx, groupID, user.ID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.NewValidationError("All users in split must be members of the group")
		}

		// Calculate amount from percentage
		amount := req.Amount.Mul(splitReq.Percentage).Div(decimal.NewFromInt(100)).Round(2)

		splits = append(splits, &models.ExpenseSplit{
			UserID:     user.ID,
			Amount:     amount,
			Percentage: splitReq.Percentage,
			User:       user,
		})

		totalPercentage = totalPercentage.Add(splitReq.Percentage)
	}

	// Validate that percentages sum to 100
	if !totalPercentage.Equal(decimal.NewFromInt(100)) {
		return nil, errors.NewInvalidSplitError("Percentages must sum to 100")
	}

	return splits, nil
}

// updateBalancesAfterExpense updates user balances after creating an expense
func (s *expenseService) updateBalancesAfterExpense(ctx context.Context, tx *database.Tx, expense *models.Expense, splits []*models.ExpenseSplit) error {
	// For each split, increase the user's debt (positive balance means they owe money)
	for _, split := range splits {
		err := s.balanceRepo.UpdateBalance(ctx, tx, expense.GroupID, split.UserID, split.Amount, expense.Currency)
		if err != nil {
			return err
		}
	}

	// Decrease the payer's debt (they paid for others)
	err := s.balanceRepo.UpdateBalance(ctx, tx, expense.GroupID, expense.PaidBy, expense.Amount.Neg(), expense.Currency)
	if err != nil {
		return err
	}

	return nil
}

// ListExpenses retrieves expenses with filtering
func (s *expenseService) ListExpenses(ctx context.Context, filter *models.ExpenseFilter) (*models.ExpenseListResponse, error) {
	expenses, total, err := s.expenseRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list expenses", zap.Error(err))
		return nil, err
	}

	// Get splits for each expense
	for _, expense := range expenses {
		expense.Splits, err = s.expenseRepo.GetExpenseSplits(ctx, expense.ID)
		if err != nil {
			return nil, err
		}
	}

	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return &models.ExpenseListResponse{
		Expenses:   expenses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

// GetGroupExpenses retrieves expenses for a specific group
func (s *expenseService) GetGroupExpenses(ctx context.Context, groupUUID string, page, limit int) ([]*models.Expense, error) {
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

	expenses, err := s.expenseRepo.GetGroupExpenses(ctx, group.ID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get group expenses", zap.Error(err), zap.String("groupUUID", groupUUID))
		return nil, err
	}

	// Get splits for each expense
	for _, expense := range expenses {
		expense.Splits, err = s.expenseRepo.GetExpenseSplits(ctx, expense.ID)
		if err != nil {
			return nil, err
		}
	}

	return expenses, nil
}

// GetUserExpenses retrieves expenses paid by a specific user
func (s *expenseService) GetUserExpenses(ctx context.Context, userUUID string, page, limit int) ([]*models.Expense, error) {
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

	expenses, err := s.expenseRepo.GetUserExpenses(ctx, user.ID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get user expenses", zap.Error(err), zap.String("userUUID", userUUID))
		return nil, err
	}

	// Get splits for each expense
	for _, expense := range expenses {
		expense.Splits, err = s.expenseRepo.GetExpenseSplits(ctx, expense.ID)
		if err != nil {
			return nil, err
		}
	}

	return expenses, nil
}
