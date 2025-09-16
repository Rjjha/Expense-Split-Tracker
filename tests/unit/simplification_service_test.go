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

type MockBalanceRepository3 struct{ mock.Mock }

type MockGroupRepository3 struct{ mock.Mock }

type MockSettlementRepository3 struct{ mock.Mock }

type MockUserRepository3 struct{ mock.Mock }

type MockDB3 struct{ mock.Mock }

// BalanceRepository methods
func (m *MockBalanceRepository3) Upsert(ctx context.Context, tx *database.Tx, balance *models.Balance) error {
	return nil
}
func (m *MockBalanceRepository3) GetByGroupAndUser(ctx context.Context, groupID, userID int64, currency string) (*models.Balance, error) {
	return nil, nil
}
func (m *MockBalanceRepository3) GetGroupBalances(ctx context.Context, groupID int64, currency string) ([]*models.Balance, error) {
	args := m.Called(ctx, groupID, currency)
	return args.Get(0).([]*models.Balance), args.Error(1)
}
func (m *MockBalanceRepository3) GetUserBalances(ctx context.Context, userID int64) ([]*models.Balance, error) {
	return nil, nil
}
func (m *MockBalanceRepository3) UpdateBalance(ctx context.Context, tx *database.Tx, groupID, userID int64, amount decimal.Decimal, currency string) error {
	return nil
}

// GroupRepository methods
func (m *MockGroupRepository3) Create(ctx context.Context, tx *database.Tx, group *models.Group) error {
	return nil
}
func (m *MockGroupRepository3) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository3) GetByUUID(ctx context.Context, uuid string) (*models.Group, error) {
	args := m.Called(ctx, uuid)
	return args.Get(0).(*models.Group), args.Error(1)
}
func (m *MockGroupRepository3) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository3) GetUserGroups(ctx context.Context, userID int64, offset, limit int) ([]*models.Group, error) {
	return nil, nil
}
func (m *MockGroupRepository3) AddMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	return nil
}
func (m *MockGroupRepository3) RemoveMember(ctx context.Context, tx *database.Tx, groupID, userID int64) error {
	return nil
}
func (m *MockGroupRepository3) GetMembers(ctx context.Context, groupID int64) ([]*models.User, error) {
	return nil, nil
}
func (m *MockGroupRepository3) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return true, nil
}

// SettlementRepository methods
func (m *MockSettlementRepository3) Create(ctx context.Context, tx *database.Tx, settlement *models.Settlement) error {
	return nil
}
func (m *MockSettlementRepository3) GetByID(ctx context.Context, id int64) (*models.Settlement, error) {
	return nil, nil
}
func (m *MockSettlementRepository3) GetByUUID(ctx context.Context, uuid string) (*models.Settlement, error) {
	return nil, nil
}
func (m *MockSettlementRepository3) List(ctx context.Context, filter *models.SettlementFilter) ([]*models.Settlement, int, error) {
	return nil, 0, nil
}
func (m *MockSettlementRepository3) GetGroupSettlements(ctx context.Context, groupID int64, offset, limit int) ([]*models.Settlement, error) {
	return nil, nil
}
func (m *MockSettlementRepository3) GetUserSettlements(ctx context.Context, userID int64, offset, limit int) ([]*models.Settlement, error) {
	return nil, nil
}

// UserRepository methods
func (m *MockUserRepository3) Create(ctx context.Context, tx *database.Tx, user *models.User) error {
	return nil
}
func (m *MockUserRepository3) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository3) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository3) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository3) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	return nil, nil
}

// DBTransactor
func (m *MockDB3) WithTransaction(fn func(tx *database.Tx) error) error { return nil }

func TestSettlementService_SimplifyDebts_GeneratesSuggestions(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	group := &models.Group{ID: 10, UUID: "11111111-1111-1111-1111-111111111111"}
	alice := &models.User{ID: 1, UUID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: "Alice"}
	bob := &models.User{ID: 2, UUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "Bob"}
	carol := &models.User{ID: 3, UUID: "cccccccc-cccc-cccc-cccc-cccccccccccc", Name: "Carol"}

	br := new(MockBalanceRepository3)
	gr := new(MockGroupRepository3)
	sr := new(MockSettlementRepository3)
	ur := new(MockUserRepository3)
	db := new(MockDB3)

	gr.On("GetByUUID", mock.Anything, group.UUID).Return(group, nil)
	br.On("GetGroupBalances", mock.Anything, group.ID, "USD").Return([]*models.Balance{
		{GroupID: group.ID, UserID: alice.ID, User: alice, Balance: decimal.NewFromInt(50)},  // owes 50
		{GroupID: group.ID, UserID: bob.ID, User: bob, Balance: decimal.NewFromInt(-30)},     // owed 30
		{GroupID: group.ID, UserID: carol.ID, User: carol, Balance: decimal.NewFromInt(-20)}, // owed 20
	}, nil)

	settlementSvc := service.NewSettlementService(sr, gr, ur, br, db, logger)

	result, err := settlementSvc.SimplifyDebts(ctx, group.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Suggestions))
	// Ensure suggestions sum equals total owed
	total := decimal.Zero
	for _, s := range result.Suggestions {
		total = total.Add(s.Amount)
	}
	assert.True(t, total.Equal(decimal.NewFromInt(50)))
	assert.GreaterOrEqual(t, result.OriginalTransactions, result.SimplifiedTransactions)
}
