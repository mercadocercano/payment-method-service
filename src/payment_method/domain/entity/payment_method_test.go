package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPaymentMethod_WithValidParams_CreatesInstance(t *testing.T) {
	// Arrange
	code := "cash"
	name := "Efectivo"
	desc := "Pago en efectivo"

	// Act
	pm := NewPaymentMethod(nil, code, name, &desc, true)

	// Assert
	require.NotNil(t, pm)
	assert.NotEqual(t, uuid.Nil, pm.ID)
	assert.Nil(t, pm.TenantID)
	assert.Equal(t, code, pm.Code)
	assert.Equal(t, name, pm.Name)
	assert.Equal(t, &desc, pm.Description)
	assert.True(t, pm.IsActive)
	assert.False(t, pm.CreatedAt.IsZero())
	assert.False(t, pm.UpdatedAt.IsZero())
}

func TestNewPaymentMethod_WithTenantID_SetsTenantID(t *testing.T) {
	// Arrange
	tenantID := uuid.New()

	// Act
	pm := NewPaymentMethod(&tenantID, "debit_card", "Tarjeta de Debito", nil, true)

	// Assert
	require.NotNil(t, pm)
	require.NotNil(t, pm.TenantID)
	assert.Equal(t, tenantID, *pm.TenantID)
}

func TestNewPaymentMethod_WithNilDescription_SetsNilDescription(t *testing.T) {
	// Act
	pm := NewPaymentMethod(nil, "cash", "Efectivo", nil, true)

	// Assert
	assert.Nil(t, pm.Description)
}

func TestNewPaymentMethod_WithInactiveFlag_SetsIsActiveFalse(t *testing.T) {
	// Act
	pm := NewPaymentMethod(nil, "crypto", "Cripto", nil, false)

	// Assert
	assert.False(t, pm.IsActive)
}

func TestIsGlobal_WithNilTenantID_ReturnsTrue(t *testing.T) {
	// Arrange
	pm := globalCashMethod()

	// Act & Assert
	assert.True(t, pm.IsGlobal())
}

func TestIsGlobal_WithTenantID_ReturnsFalse(t *testing.T) {
	// Arrange
	tenantID := uuid.New()
	pm := tenantDebitCardMethod(tenantID)

	// Act & Assert
	assert.False(t, pm.IsGlobal())
}

func TestNewPaymentMethod_GeneratesUniqueIDs(t *testing.T) {
	// Act
	pm1 := NewPaymentMethod(nil, "cash", "Efectivo", nil, true)
	pm2 := NewPaymentMethod(nil, "cash", "Efectivo", nil, true)

	// Assert
	assert.NotEqual(t, pm1.ID, pm2.ID)
}

func TestNewPaymentMethod_SetsCreatedAtAndUpdatedAtEqual(t *testing.T) {
	// Act
	pm := NewPaymentMethod(nil, "cash", "Efectivo", nil, true)

	// Assert
	assert.Equal(t, pm.CreatedAt, pm.UpdatedAt)
}
