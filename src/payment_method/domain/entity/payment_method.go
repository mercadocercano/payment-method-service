package entity

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod representa un método de pago en el sistema (read-only para MVP POS)
type PaymentMethod struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    *uuid.UUID `json:"tenant_id"`  // NULL = global, NOT NULL = específico de tenant
	Code        string     `json:"code"`       // "cash", "debit_card", etc.
	Name        string     `json:"name"`       // "Efectivo", "Tarjeta de Débito", etc.
	Description *string    `json:"description"` // Opcional
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewPaymentMethod crea una nueva instancia de PaymentMethod (para seeds/futuro)
func NewPaymentMethod(
	tenantID *uuid.UUID,
	code string,
	name string,
	description *string,
	isActive bool,
) *PaymentMethod {
	now := time.Now()
	return &PaymentMethod{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsGlobal retorna true si el método de pago es global (tenant_id = NULL)
func (pm *PaymentMethod) IsGlobal() bool {
	return pm.TenantID == nil
}
