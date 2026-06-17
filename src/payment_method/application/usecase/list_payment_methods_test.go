package usecase

import (
	"errors"
	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/domain/port/mocks"
	"payment_method/src/payment_method/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPaymentMethods_WithValidTenantID_ReturnsMethods(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()
	methods := []*entity.PaymentMethod{
		testutil.GlobalCashMethod(),
		testutil.TenantDebitCardMethod(tenantID),
	}

	mockRepo.On("FindAll", tenantID, true).Return(methods, nil)

	// Act
	result, err := uc.Execute(tenantID, true)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Items, 2)
	mockRepo.AssertExpectations(t)
}

func TestListPaymentMethods_WithActiveOnlyTrue_PassesFilterToRepo(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()

	mockRepo.On("FindAll", tenantID, true).Return([]*entity.PaymentMethod{
		testutil.GlobalCashMethod(),
	}, nil)

	// Act
	result, err := uc.Execute(tenantID, true)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalCount)
	mockRepo.AssertExpectations(t)
}

func TestListPaymentMethods_WithActiveOnlyFalse_IncludesInactiveMethods(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()
	methods := []*entity.PaymentMethod{
		testutil.GlobalCashMethod(),
		testutil.InactiveMethod(),
	}

	mockRepo.On("FindAll", tenantID, false).Return(methods, nil)

	// Act
	result, err := uc.Execute(tenantID, false)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Items, 2)
	mockRepo.AssertExpectations(t)
}

func TestListPaymentMethods_WithNilTenantID_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	// Act
	result, err := uc.Execute(uuid.Nil, true)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "tenant_id is required", err.Error())
	mockRepo.AssertNotCalled(t, "FindAll")
}

func TestListPaymentMethods_WithRepositoryError_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()
	repoErr := errors.New("database connection failed")

	mockRepo.On("FindAll", tenantID, true).Return(nil, repoErr)

	// Act
	result, err := uc.Execute(tenantID, true)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection failed", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestListPaymentMethods_WithEmptyResult_ReturnsEmptyList(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()

	mockRepo.On("FindAll", tenantID, true).Return([]*entity.PaymentMethod{}, nil)

	// Act
	result, err := uc.Execute(tenantID, true)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.TotalCount)
	assert.Empty(t, result.Items)
	mockRepo.AssertExpectations(t)
}

func TestListPaymentMethods_MapsIsGlobalCorrectly(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewListPaymentMethodsUseCase(mockRepo, nil)

	tenantID := uuid.New()
	methods := []*entity.PaymentMethod{
		testutil.GlobalCashMethod(),
		testutil.TenantDebitCardMethod(tenantID),
	}

	mockRepo.On("FindAll", tenantID, true).Return(methods, nil)

	// Act
	result, err := uc.Execute(tenantID, true)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.Items[0].IsGlobal, "global method should have IsGlobal=true")
	assert.False(t, result.Items[1].IsGlobal, "tenant method should have IsGlobal=false")
	mockRepo.AssertExpectations(t)
}
