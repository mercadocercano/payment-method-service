package testutil

import (
	"payment_method/src/payment_method/domain/entity"
	"time"

	"github.com/google/uuid"
)

// PaymentMethodMother es un Object Mother para crear instancias de PaymentMethod en tests
type PaymentMethodMother struct {
	id          uuid.UUID
	tenantID    *uuid.UUID
	code        string
	name        string
	description *string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewPaymentMethodMother crea un mother con valores por defecto validos
func NewPaymentMethodMother() *PaymentMethodMother {
	now := time.Now().Truncate(time.Second)
	return &PaymentMethodMother{
		id:        uuid.New(),
		tenantID:  nil,
		code:      "cash",
		name:      "Efectivo",
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}
}

func (m *PaymentMethodMother) WithID(id uuid.UUID) *PaymentMethodMother {
	m.id = id
	return m
}

func (m *PaymentMethodMother) WithTenantID(tenantID *uuid.UUID) *PaymentMethodMother {
	m.tenantID = tenantID
	return m
}

func (m *PaymentMethodMother) WithCode(code string) *PaymentMethodMother {
	m.code = code
	return m
}

func (m *PaymentMethodMother) WithName(name string) *PaymentMethodMother {
	m.name = name
	return m
}

func (m *PaymentMethodMother) WithDescription(description *string) *PaymentMethodMother {
	m.description = description
	return m
}

func (m *PaymentMethodMother) WithIsActive(isActive bool) *PaymentMethodMother {
	m.isActive = isActive
	return m
}

func (m *PaymentMethodMother) Build() *entity.PaymentMethod {
	return &entity.PaymentMethod{
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

// GlobalCashMethod crea un metodo de pago global en efectivo
func GlobalCashMethod() *entity.PaymentMethod {
	return NewPaymentMethodMother().
		WithCode("cash").
		WithName("Efectivo").
		Build()
}

// TenantDebitCardMethod crea un metodo de pago de debito para un tenant especifico
func TenantDebitCardMethod(tenantID uuid.UUID) *entity.PaymentMethod {
	return NewPaymentMethodMother().
		WithTenantID(&tenantID).
		WithCode("debit_card").
		WithName("Tarjeta de Debito").
		Build()
}

// InactiveMethod crea un metodo de pago inactivo
func InactiveMethod() *entity.PaymentMethod {
	return NewPaymentMethodMother().
		WithCode("crypto").
		WithName("Criptomonedas").
		WithIsActive(false).
		Build()
}
