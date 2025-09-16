package unit

import (
	"context"
	"strings"
	"testing"

	"expense-split-tracker/internal/database"
	"expense-split-tracker/internal/models"
	"expense-split-tracker/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockSettlementRepository struct{ mock.Mock }

type MockBalanceRepository2 struct{ mock.Mock }

type MockGroupRepository2 struct{ mock.Mock }

type MockUserRepository2 struct{ mock.Mock }

type MockDB2 struct{ mock.Mock }

func (m *MockSettlementRepository) Create(ctx context.Context, tx *database.Tx, settlement *models.Settlement) error {
	args := m.Called(ctx, tx, settlement)
	return args.Error(0)
}

func (m *MockSettlementRepository) GetByID(ctx context.Context, id int64) (*models.Settlement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Settlement), args.Error(1)
}

func (m *MockSettlementRepository) GetByUUID(ctx context.Context, uuid string) (*models.Settlement, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Settlement), args.Error(1)
}

func (m *MockSettlementRepository) List(ctx context.Context, filter *models.SettlementFilter) ([]*models.Settlement, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Settlement), args.Int(1), args.Error(2)
}

func (m *MockSettlementRepository) GetGroupSettlements(ctx context.Context, groupID int64, offset, limit int) ([]*models.Settlement, error) {
	args := m.Called(ctx, groupID, offset, limit)
	return args.Get(0).([]*models.Settlement), args.Error(1)
}

func (m *MockSettlementRepository) GetUserSettlements(ctx context.Context, userID int64, offset, limit int) ([]*models.Settlement, error) {
	args := m.Called(ctx, userID, offset, limit)
	return args.Get(0).([]*models.Settlement), args.Error(1)
}

func (m *MockBalanceRepository2) UpdateBalance(ctx context.Context, tx *database.Tx, groupID, userID int64, amount decimal.Decimal, currency string) error {
	args := m.Called(ctx, tx, groupID, userID, amount, currency)
	return args.Error(0)
}

func (m *MockBalanceRepository2) Upsert(ctx context.Context, tx *database.Tx, balance *models.Balance) error {
	return nil
}
func (m *MockBalanceRepository2) GetByGroupAndUser(ctx context.Context, groupID, userID int64, currency string) (*models.Balance, error) {
	args := m.Called(ctx, groupID, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Balance), args.Error(1)
}
func (m *MockBalanceRepository2) GetGroupBalances(ctx context.Context, groupID int64, currency string) ([]*models.Balance, error) {
	return nil, nil
}
func (m *MockBalanceRepository2) GetUserBalances(ctx context.Context, userID int64) ([]*models.Balance, error) {
	return nil, nil
}

func (m *MockGroupRepository2) GetByUUID(ctx context.Context, uuid string) (*models.Group, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}
func (m *MockGroupRepository2) Create(ctx context.Context, tx *database.Tx, group *models.Group) error {
	return nil
}
func (m *MockGroupRepository2) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository2) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository2) GetUserGroups(ctx context.Context, userID int64, offset, limit int) ([]*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository2) AddMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	return nil
}
func (m *MockGroupRepository2) RemoveMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	return nil
}
func (m *MockGroupRepository2) GetMembers(ctx context.Context, groupID int64) ([]*models.User, error) {
	return nil, nil
}
func (m *MockGroupRepository2) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository2) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository2) Create(ctx context.Context, tx *database.Tx, user *models.User) error {
	return nil
}
func (m *MockUserRepository2) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository2) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository2) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	return nil, nil
}

func (m *MockDB2) WithTransaction(fn func(tx *database.Tx) error) error {
	args := m.Called(fn)
	if err := fn(nil); err != nil {
		return err
	}
	return args.Error(0)
}

func TestSettlementService_CreateSettlement_Success(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	settlementRepo := new(MockSettlementRepository)
	groupRepo := new(MockGroupRepository2)
	userRepo := new(MockUserRepository2)
	balanceRepo := new(MockBalanceRepository2)
	db := new(MockDB2)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111"}
	fromUser := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	toUser := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}
	currency := "USD"

	groupRepo.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	userRepo.On("GetByUUID", mock.Anything, fromUser.UUID).Return(fromUser, nil)
	userRepo.On("GetByUUID", mock.Anything, toUser.UUID).Return(toUser, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, fromUser.ID).Return(true, nil)
	groupRepo.On("IsMember", mock.Anything, group.ID, toUser.ID).Return(true, nil)
	balanceRepo.On("GetByGroupAndUser", mock.Anything, group.ID, fromUser.ID, currency).Return(&models.Balance{GroupID: group.ID, UserID: fromUser.ID, Balance: decimal.NewFromInt(100), Currency: currency}, nil)

	settlementRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Settlement")).Return(nil)
	settlementRepo.On("GetByUUID", mock.Anything, mock.AnythingOfType("string")).Return(&models.Settlement{}, nil)

	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, fromUser.ID, decimal.NewFromInt(50).Neg(), currency).Return(nil)
	balanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, group.ID, toUser.ID, decimal.NewFromInt(50), currency).Return(nil)

	db.On("WithTransaction", mock.AnythingOfType("func(*database.Tx) error")).Return(nil)

	s := service.NewSettlementService(settlementRepo, groupRepo, userRepo, balanceRepo, db, logger)

	res, err := s.CreateSettlement(ctx, &models.CreateSettlementRequest{
		GroupUUID:    group.UUID,
		FromUserUUID: fromUser.UUID,
		ToUserUUID:   toUser.UUID,
		Amount:       decimal.NewFromInt(50),
		Currency:     currency,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestSettlementService_CreateSettlement_AmountExceedsDebt(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	sr := new(MockSettlementRepository)
	gr := new(MockGroupRepository2)
	ur := new(MockUserRepository2)
	br := new(MockBalanceRepository2)
	db := new(MockDB2)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111"}
	fromUser := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	toUser := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}

	gr.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	ur.On("GetByUUID", mock.Anything, fromUser.UUID).Return(fromUser, nil)
	ur.On("GetByUUID", mock.Anything, toUser.UUID).Return(toUser, nil)
	gr.On("IsMember", mock.Anything, group.ID, fromUser.ID).Return(true, nil)
	gr.On("IsMember", mock.Anything, group.ID, toUser.ID).Return(true, nil)
	br.On("GetByGroupAndUser", mock.Anything, group.ID, fromUser.ID, "USD").Return(&models.Balance{GroupID: group.ID, UserID: fromUser.ID, Balance: decimal.NewFromInt(20), Currency: "USD"}, nil)

	s := service.NewSettlementService(sr, gr, ur, br, db, logger)

	res, err := s.CreateSettlement(ctx, &models.CreateSettlementRequest{
		GroupUUID:    group.UUID,
		FromUserUUID: fromUser.UUID,
		ToUserUUID:   toUser.UUID,
		Amount:       decimal.NewFromInt(50),
		Currency:     "USD",
	})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.True(t, strings.Contains(strings.ToLower(err.Error()), "insufficient"))
}

func TestSettlementService_CreateSettlement_SameUser(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	s := service.NewSettlementService(new(MockSettlementRepository), new(MockGroupRepository2), new(MockUserRepository2), new(MockBalanceRepository2), new(MockDB2), logger)

	res, err := s.CreateSettlement(ctx, &models.CreateSettlementRequest{
		GroupUUID:    "11111111-1111-1111-1111-111111111111",
		FromUserUUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		ToUserUUID:   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Amount:       decimal.NewFromInt(10),
		Currency:     "USD",
	})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "cannot be the same")
}
