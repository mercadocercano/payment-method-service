package port

import (
	"context"
	"payment_method/src/payment_method/domain/entity"

	"github.com/google/uuid"
)

// PaymentMethodRepository define el contrato para el repositorio de métodos de pago.
// ctx es obligatorio: el adapter Postgres fija el contexto RLS (app.tenant_id) por
// transacción (RULE-10), y ese contexto se propaga vía ctx desde el request autenticado.
type PaymentMethodRepository interface {
	// FindByID busca un método de pago habilitado del tenant por su ID
	FindByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*entity.PaymentMethod, error)

	// FindAll retorna los métodos del plano tenant (mapping) con code/name del catálogo global
	FindAll(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error)
}
