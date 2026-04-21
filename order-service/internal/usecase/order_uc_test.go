package usecase

import (
	"context"
	"errors"
	"testing"

	"ecommerce/order-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==========================================
// 1. CREATE THE MOCKS (Fake Dependencies)
// ==========================================

// MockOrderRepo acts like our Postgres database, but in memory!
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) Create(ctx context.Context, o *domain.Order) (int64, error) {
	args := m.Called(ctx, o)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *MockOrderRepo) GetStatus(ctx context.Context, id int64) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

// MockKafkaPub acts like our Kafka Writer
type MockKafkaPub struct {
	mock.Mock
}

func (m *MockKafkaPub) PublishEvent(ctx context.Context, topic string, eventType string, event domain.SagaEvent) error {
	args := m.Called(ctx, topic, eventType, event)
	return args.Error(0)
}

// ==========================================
// 2. WRITE THE TEST CASES
// ==========================================

func TestCheckout_Success(t *testing.T) {
	// Setup our mocks
	mockRepo := new(MockOrderRepo)
	mockPub := new(MockKafkaPub)

	// Create the UseCase with the FAKE dependencies
	uc := NewOrderUseCase(mockRepo, mockPub)

	ctx := context.Background()
	userID := int64(1)
	productID := int64(99)
	amount := 49.99
	expectedOrderID := int64(1001)

	// TELL THE MOCKS HOW TO BEHAVE:
	// "When Create is called, return ID 1001 and NO error"
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(expectedOrderID, nil)
	
	// "When PublishEvent is called, return NO error"
	mockPub.On("PublishEvent", ctx, "order-events", "OrderCreated", mock.Anything).Return(nil)

	// ACT: Run the actual function
	orderID, err := uc.Checkout(ctx, userID, productID, amount)

	// ASSERT: Check the results
	assert.NoError(t, err)              // We expect no error
	assert.Equal(t, expectedOrderID, orderID) // We expect the ID to be 1001

	// Verify that the UseCase actually called our Fakes!
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestCheckout_DatabaseFailure(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockPub := new(MockKafkaPub)
	uc := NewOrderUseCase(mockRepo, mockPub)

	ctx := context.Background()

	// TELL THE MOCK DATABASE TO FAIL
	mockRepo.On("Create", ctx, mock.Anything).Return(int64(0), errors.New("database down"))

	// ACT
	orderID, err := uc.Checkout(ctx, 1, 99, 49.99)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, int64(0), orderID)
	assert.Equal(t, "database down", err.Error())

	// Ensure Kafka was NEVER called because the DB failed first!
	mockPub.AssertNotCalled(t, "PublishEvent")
}