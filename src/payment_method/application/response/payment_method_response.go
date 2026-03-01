package response

import (
	"payment_method/src/payment_method/domain/entity"
	"time"

	"github.com/google/uuid"
)

// PaymentMethodResponse es el DTO de respuesta para un método de pago
type PaymentMethodResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    *uuid.UUID `json:"tenant_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	IsActive    bool       `json:"is_active"`
	IsGlobal    bool       `json:"is_global"` // Computed field para conveniencia del cliente
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FromEntity convierte una entidad PaymentMethod a PaymentMethodResponse
func FromEntity(pm *entity.PaymentMethod) *PaymentMethodResponse {
	return &PaymentMethodResponse{
		ID:          pm.ID,
		TenantID:    pm.TenantID,
		Code:        pm.Code,
		Name:        pm.Name,
		Description: pm.Description,
		IsActive:    pm.IsActive,
		IsGlobal:    pm.IsGlobal(),
		CreatedAt:   pm.CreatedAt,
		UpdatedAt:   pm.UpdatedAt,
	}
}

// ListPaymentMethodsResponse es el DTO de respuesta para la lista de métodos de pago
type ListPaymentMethodsResponse struct {
	Items      []*PaymentMethodResponse `json:"items"`
	TotalCount int                      `json:"total_count"`
}
