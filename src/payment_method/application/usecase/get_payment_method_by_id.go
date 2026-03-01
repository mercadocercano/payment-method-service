package usecase

import (
	"errors"
	"payment_method/src/payment_method/application/response"
	"payment_method/src/payment_method/domain/port"

	"github.com/google/uuid"
)

// GetPaymentMethodByIDUseCase maneja la obtención de un método de pago por ID
type GetPaymentMethodByIDUseCase struct {
	repository port.PaymentMethodRepository
}

// NewGetPaymentMethodByIDUseCase crea una nueva instancia del caso de uso
func NewGetPaymentMethodByIDUseCase(repository port.PaymentMethodRepository) *GetPaymentMethodByIDUseCase {
	return &GetPaymentMethodByIDUseCase{
		repository: repository,
	}
}

// Execute ejecuta el caso de uso
func (uc *GetPaymentMethodByIDUseCase) Execute(id uuid.UUID, tenantID uuid.UUID) (*response.PaymentMethodResponse, error) {
	// Validar que el tenant ID no esté vacío
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}

	// Buscar el método de pago
	paymentMethod, err := uc.repository.FindByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	if paymentMethod == nil {
		return nil, errors.New("payment method not found")
	}

	// Convertir a response
	return response.FromEntity(paymentMethod), nil
}
