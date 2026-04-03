package usecase

import (
	"errors"
	"payment_method/src/payment_method/domain/port/mocks"
	"payment_method/src/payment_method/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaymentMethodByID_WithValidID_ReturnsPaymentMethod(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	tenantID := uuid.New()
	pm := testutil.GlobalCashMethod()

	mockRepo.On("FindByID", pm.ID, tenantID).Return(pm, nil)

	// Act
	result, err := uc.Execute(pm.ID, tenantID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, pm.ID, result.ID)
	assert.Equal(t, pm.Code, result.Code)
	assert.Equal(t, pm.Name, result.Name)
	assert.Equal(t, pm.IsActive, result.IsActive)
	assert.True(t, result.IsGlobal)
	mockRepo.AssertExpectations(t)
}

func TestGetPaymentMethodByID_WithTenantSpecificMethod_ReturnsNonGlobal(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	tenantID := uuid.New()
	pm := testutil.TenantDebitCardMethod(tenantID)

	mockRepo.On("FindByID", pm.ID, tenantID).Return(pm, nil)

	// Act
	result, err := uc.Execute(pm.ID, tenantID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsGlobal)
	assert.Equal(t, &tenantID, result.TenantID)
	mockRepo.AssertExpectations(t)
}

func TestGetPaymentMethodByID_WithNilTenantID_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	// Act
	result, err := uc.Execute(uuid.New(), uuid.Nil)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "tenant_id is required", err.Error())
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestGetPaymentMethodByID_WithNonExistentID_ReturnsNotFoundError(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	tenantID := uuid.New()
	nonExistentID := uuid.New()

	mockRepo.On("FindByID", nonExistentID, tenantID).Return(nil, nil)

	// Act
	result, err := uc.Execute(nonExistentID, tenantID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "payment method not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetPaymentMethodByID_WithRepositoryError_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	tenantID := uuid.New()
	pmID := uuid.New()
	repoErr := errors.New("database connection failed")

	mockRepo.On("FindByID", pmID, tenantID).Return(nil, repoErr)

	// Act
	result, err := uc.Execute(pmID, tenantID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection failed", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetPaymentMethodByID_WithDescription_MapsDescriptionCorrectly(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPaymentMethodRepository)
	uc := NewGetPaymentMethodByIDUseCase(mockRepo)

	tenantID := uuid.New()
	desc := "Pago en efectivo en moneda local"
	pm := testutil.NewPaymentMethodMother().
		WithDescription(&desc).
		Build()

	mockRepo.On("FindByID", pm.ID, tenantID).Return(pm, nil)

	// Act
	result, err := uc.Execute(pm.ID, tenantID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result.Description)
	assert.Equal(t, desc, *result.Description)
	mockRepo.AssertExpectations(t)
}
