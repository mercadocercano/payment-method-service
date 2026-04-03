package entity

import (
	"time"

	"github.com/google/uuid"
)

// paymentMethodMother es un Object Mother interno para tests del package entity
type paymentMethodMother struct {
	id          uuid.UUID
	tenantID    *uuid.UUID
	code        string
	name        string
	description *string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

func newMother() *paymentMethodMother {
	now := time.Now().Truncate(time.Second)
	return &paymentMethodMother{
		id:        uuid.New(),
		tenantID:  nil,
		code:      "cash",
		name:      "Efectivo",
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}
}

func (m *paymentMethodMother) withTenantID(tenantID *uuid.UUID) *paymentMethodMother {
	m.tenantID = tenantID
	return m
}

func (m *paymentMethodMother) withCode(code string) *paymentMethodMother {
	m.code = code
	return m
}

func (m *paymentMethodMother) withName(name string) *paymentMethodMother {
	m.name = name
	return m
}

func (m *paymentMethodMother) withIsActive(isActive bool) *paymentMethodMother {
	m.isActive = isActive
	return m
}

func (m *paymentMethodMother) build() *PaymentMethod {
	return &PaymentMethod{
		ID:          m.id,
		TenantID:    m.tenantID,
		Code:        m.code,
		Name:        m.name,
		Description: m.description,
		IsActive:    m.isActive,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

func globalCashMethod() *PaymentMethod {
	return newMother().build()
}

func tenantDebitCardMethod(tenantID uuid.UUID) *PaymentMethod {
	return newMother().
		withTenantID(&tenantID).
		withCode("debit_card").
		withName("Tarjeta de Debito").
		build()
}
