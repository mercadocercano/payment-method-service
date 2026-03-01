package port

import (
	"payment_method/src/payment_method/domain/entity"
	"github.com/google/uuid"
)

// PaymentMethodRepository define el contrato para el repositorio de métodos de pago
type PaymentMethodRepository interface {
	// FindByID busca un método de pago por su ID (global o del tenant)
	FindByID(id uuid.UUID, tenantID uuid.UUID) (*entity.PaymentMethod, error)

	// FindAll retorna todos los métodos de pago disponibles para un tenant
	// (incluye métodos globales + específicos del tenant)
	FindAll(tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error)
}
