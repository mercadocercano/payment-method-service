package usecase

import (
	"errors"
	"payment_method/src/payment_method/application/response"
	"payment_method/src/payment_method/domain/port"

	"github.com/google/uuid"
)

// ListPaymentMethodsUseCase maneja la lista de métodos de pago
type ListPaymentMethodsUseCase struct {
	repository port.PaymentMethodRepository
}

// NewListPaymentMethodsUseCase crea una nueva instancia del caso de uso
func NewListPaymentMethodsUseCase(repository port.PaymentMethodRepository) *ListPaymentMethodsUseCase {
	return &ListPaymentMethodsUseCase{
		repository: repository,
	}
}

// Execute ejecuta el caso de uso
func (uc *ListPaymentMethodsUseCase) Execute(tenantID uuid.UUID, activeOnly bool) (*response.ListPaymentMethodsResponse, error) {
	// Validar que el tenant ID no esté vacío
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}

	// Obtener métodos de pago del repositorio
	paymentMethods, err := uc.repository.FindAll(tenantID, activeOnly)
	if err != nil {
		return nil, err
	}

	// Convertir a response
	items := make([]*response.PaymentMethodResponse, 0, len(paymentMethods))
	for _, pm := range paymentMethods {
		items = append(items, response.FromEntity(pm))
	}

	return &response.ListPaymentMethodsResponse{
		Items:      items,
		TotalCount: len(items),
	}, nil
}
