package unit

import (
	"context"
	"testing"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// Mocks for Expense service dependencies

type MockExpenseRepositoryES struct{ mock.Mock }

type MockGroupRepositoryES struct{ mock.Mock }

type MockUserRepositoryES struct{ mock.Mock }

type MockBalanceRepositoryES struct{ mock.Mock }

type MockDBES struct{ mock.Mock }

func (m *MockExpenseRepositoryES) Create(ctx context.Context, tx *database.Tx, expense *models.Expense) error {
	args := m.Called(ctx, tx, expense)
	// simulate DB auto-increment id after create
	if args.Error(0) == nil && expense.ID == 0 {
		expense.ID = 1
	}
	return args.Error(0)
}

func (m *MockExpenseRepositoryES) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Expense), args.Error(1)
}

func (m *MockExpenseRepositoryES) List(ctx context.Context, filter *models.ExpenseFilter) ([]*models.Expense, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Expense), args.Int(1), args.Error(2)
}

func (m *MockExpenseRepositoryES) GetGroupExpenses(ctx context.Context, groupID int64, offset, limit int) ([]*models.Expense, error) {
	args := m.Called(ctx, groupID, offset, limit)
	return args.Get(0).([]*models.Expense), args.Error(1)
}

func (m *MockExpenseRepositoryES) GetUserExpenses(ctx context.Context, userID int64, offset, limit int) ([]*models.Expense, error) {
	args := m.Called(ctx, userID, offset, limit)
	return args.Get(0).([]*models.Expense), args.Error(1)
}

func (m *MockExpenseRepositoryES) CreateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error {
	args := m.Called(ctx, tx, split)
	return args.Error(0)
}

func (m *MockExpenseRepositoryES) GetExpenseSplits(ctx context.Context, expenseID int64) ([]*models.ExpenseSplit, error) {
	args := m.Called(ctx, expenseID)
	return args.Get(0).([]*models.ExpenseSplit), args.Error(1)
}

func (m *MockExpenseRepositoryES) UpdateSplit(ctx context.Context, tx *database.Tx, split *models.ExpenseSplit) error {
	args := m.Called(ctx, tx, split)
	return args.Error(0)
}

func (m *MockGroupRepositoryES) Create(ctx context.Context, tx *database.Tx, group *models.Group) error {
	args := m.Called(ctx, tx, group)
	return args.Error(0)
}

func (m *MockGroupRepositoryES) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepositoryES) GetByUUID(ctx context.Context, uuid string) (*models.Group, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepositoryES) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepositoryES) GetUserGroups(ctx context.Context, userID int64, offset, limit int) ([]*models.Group, error) {
	args := m.Called(ctx, userID, offset, limit)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepositoryES) AddMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	args := m.Called(ctx, tx, groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepositoryES) RemoveMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	args := m.Called(ctx, tx, groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepositoryES) GetMembers(ctx context.Context, groupID int64) ([]*models.User, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockGroupRepositoryES) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepositoryES) Create(ctx context.Context, tx *database.Tx, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryES) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryES) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryES) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryES) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockBalanceRepositoryES) Upsert(ctx context.Context, tx *database.Tx, balance *models.Balance) error {
	args := m.Called(ctx, tx, balance)
	return args.Error(0)
}

func (m *MockBalanceRepositoryES) GetByGroupAndUser(ctx context.Context, groupID, userID int64, currency string) (*models.Balance, error) {
	args := m.Called(ctx, groupID, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Balance), args.Error(1)
}

func (m *MockBalanceRepositoryES) GetGroupBalances(ctx context.Context, groupID int64, currency string) ([]*models.Balance, error) {
	args := m.Called(ctx, groupID, currency)
	return args.Get(0).([]*models.Balance), args.Error(1)
}

func (m *MockBalanceRepositoryES) GetUserBalances(ctx context.Context, userID int64) ([]*models.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Balance), args.Error(1)
}

func (m *MockBalanceRepositoryES) UpdateBalance(ctx context.Context, tx *database.Tx, groupID, userID int64, amount decimal.Decimal, currency string) error {
	args := m.Called(ctx, tx, groupID, userID, amount, currency)
	return args.Error(0)
}

func (m *MockDBES) WithTransaction(fn func(tx *database.Tx) error) error {
	args := m.Called(fn)
	if err := fn(nil); err != nil {
		return err
	}
	return args.Error(0)
}

func TestExpenseService_CreateExpense_EqualSplit(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	expenseRepo := new(MockExpenseRepositoryES)
	groupRepo := new(MockGroupRepositoryES)
	userRepo := new(MockUserRepositoryES)
	balanceRepo := new(MockBalanceRepositoryES)
	db := new(MockDBES)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111", Name: "Trip"}
	payer := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: "Alice"}
	user2 := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "Bob"}
	user3 := &models.User{ID: 3, UUID: "cccccccc-cccc-cccc-cccc-cccccccccccc", Name: "Carol"}

	req := &models.CreateExpenseRequest{
		GroupUUID:   group.UUID,
		PaidByUUID:  payer.UUID,
		Amount:      decimal.NewFromInt(90),
		Currency:    "USD",
		Description: "Dinner",
		SplitType:   models.SplitTypeEqual,
		Splits: []models.CreateExpenseSplitRequest{
			{UserUUID: payer.UUID},
			{UserUUID: user2.UUID},
			{UserUUID: user3.UUID},
		},
	}

	groupRepo.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	userRepo.On("GetByUUID", mock.Anything, payer.UUID).Return(payer, nil)
	userRepo.On("GetByUUID", mock.Anything, user2.UUID).Return(user2, nil)
	userRepo.On("GetByUUID", mock.Anything, user3.UUID).Return(user3, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, payer.ID).Return(true, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, user2.ID).Return(true, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, user3.ID).Return(true, nil)

	expenseRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Expense")).Return(nil)
	expenseRepo.On("CreateSplit", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ExpenseSplit")).Return(nil).Times(3)
	expenseRepo.On("GetExpenseSplits", mock.Anything, int64(1)).Return([]*models.ExpenseSplit{
		{UserID: payer.ID, Amount: decimal.NewFromInt(30)},
		{UserID: user2.ID, Amount: decimal.NewFromInt(30)},
		{UserID: user3.ID, Amount: decimal.NewFromInt(30)},
	}, nil)

	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, payer.ID, decimal.NewFromInt(90).Neg(), "USD").Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, payer.ID, mock.Anything, "USD").Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, user2.ID, mock.Anything, "USD").Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, user3.ID, mock.Anything, "USD").Return(nil)

	db.On("WithTransaction", mock.AnythingOfType("func(*database.Tx) error")).Return(nil)

	es := service.NewExpenseService(expenseRepo, groupRepo, userRepo, balanceRepo, db, logger)

	expense, err := es.CreateExpense(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, expense)
	assert.Equal(t, models.SplitTypeEqual, expense.SplitType)
	assert.Equal(t, "USD", expense.Currency)
	assert.Equal(t, 3, len(expense.Splits))
}

func TestExpenseService_CreateExpense_ExactSplit_SumMismatch(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	expenseRepo := new(MockExpenseRepositoryES)
	groupRepo := new(MockGroupRepositoryES)
	userRepo := new(MockUserRepositoryES)
	balanceRepo := new(MockBalanceRepositoryES)
	db := new(MockDBES)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111", Name: "Trip"}
	payer := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: "Alice"}
	user2 := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "Bob"}

	req := &models.CreateExpenseRequest{
		GroupUUID:   group.UUID,
		PaidByUUID:  payer.UUID,
		Amount:      decimal.NewFromInt(100),
		Currency:    "USD",
		Description: "Cab",
		SplitType:   models.SplitTypeExact,
		Splits: []models.CreateExpenseSplitRequest{
			{UserUUID: payer.UUID, Amount: decimal.NewFromInt(30)},
			{UserUUID: user2.UUID, Amount: decimal.NewFromInt(50)}, // totals 80, mismatch
		},
	}

	groupRepo.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	userRepo.On("GetByUUID", mock.Anything, payer.UUID).Return(payer, nil)
	userRepo.On("GetByUUID", mock.Anything, user2.UUID).Return(user2, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, payer.ID).Return(true, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, user2.ID).Return(true, nil)

	es := service.NewExpenseService(expenseRepo, groupRepo, userRepo, balanceRepo, db, logger)

	expense, err := es.CreateExpense(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, expense)
	assert.Contains(t, err.Error(), "Sum of split amounts must equal")
}

func TestExpenseService_CreateExpense_Percentage_SumTo100(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	expenseRepo := new(MockExpenseRepositoryES)
	groupRepo := new(MockGroupRepositoryES)
	userRepo := new(MockUserRepositoryES)
	balanceRepo := new(MockBalanceRepositoryES)
	db := new(MockDBES)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111"}
	payer := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	user2 := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}

	req := &models.CreateExpenseRequest{
		GroupUUID:   group.UUID,
		PaidByUUID:  payer.UUID,
		Amount:      decimal.NewFromInt(200),
		Currency:    "USD",
		Description: "Hotel",
		SplitType:   models.SplitTypePercentage,
		Splits: []models.CreateExpenseSplitRequest{
			{UserUUID: payer.UUID, Percentage: decimal.NewFromInt(60)},
			{UserUUID: user2.UUID, Percentage: decimal.NewFromInt(40)},
		},
	}

	groupRepo.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	userRepo.On("GetByUUID", mock.Anything, payer.UUID).Return(payer, nil)
	userRepo.On("GetByUUID", mock.Anything, user2.UUID).Return(user2, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, payer.ID).Return(true, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, user2.ID).Return(true, nil)

	expenseRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Expense")).Return(nil)
	expenseRepo.On("CreateSplit", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ExpenseSplit")).Return(nil).Times(2)
	expenseRepo.On("GetExpenseSplits", mock.Anything, int64(1)).Return([]*models.ExpenseSplit{
		{UserID: payer.ID, Amount: decimal.NewFromInt(120)},
		{UserID: user2.ID, Amount: decimal.NewFromInt(80)},
	}, nil)

	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, payer.ID, decimal.NewFromInt(200).Neg(), "USD").Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, payer.ID, mock.Anything, "USD").Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, user2.ID, mock.Anything, "USD").Return(nil)

	db.On("WithTransaction", mock.AnythingOfType("func(*database.Tx) error")).Return(nil)

	es := service.NewExpenseService(expenseRepo, groupRepo, userRepo, balanceRepo, db, logger)

	expense, err := es.CreateExpense(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, expense)
	assert.Equal(t, models.SplitTypePercentage, expense.SplitType)
	assert.Equal(t, 2, len(expense.Splits))
}

func TestExpenseService_CreateExpense_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	es := service.NewExpenseService(new(MockExpenseRepositoryES), new(MockGroupRepositoryES), new(MockUserRepositoryES), new(MockBalanceRepositoryES), new(MockDBES), logger)

	req := &models.CreateExpenseRequest{
		GroupUUID:   "invalid",
		PaidByUUID:  "also-invalid",
		Amount:      decimal.NewFromInt(10),
		Description: "x",
		SplitType:   models.SplitTypeEqual,
		Splits:      []models.CreateExpenseSplitRequest{{UserUUID: "invalid"}},
	}

	res, err := es.CreateExpense(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Invalid value")
}
