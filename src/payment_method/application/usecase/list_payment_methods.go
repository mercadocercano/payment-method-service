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
	logger     port.PaymentEventLogger
}

// NewListPaymentMethodsUseCase crea una nueva instancia del caso de uso
func NewListPaymentMethodsUseCase(repository port.PaymentMethodRepository, logger port.PaymentEventLogger) *ListPaymentMethodsUseCase {
	return &ListPaymentMethodsUseCase{
		repository: repository,
		logger:     logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (uc *ListPaymentMethodsUseCase) logEvent(e port.PaymentEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
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
		uc.logEvent(port.PaymentEvent{
			Event:    "payment.methods_list_failed",
			TenantID: tenantID.String(),
			Reason:   err.Error(),
		})
		return nil, err
	}

	// Convertir a response
	items := make([]*response.PaymentMethodResponse, 0, len(paymentMethods))
	for _, pm := range paymentMethods {
		items = append(items, response.FromEntity(pm))
	}

	uc.logEvent(port.PaymentEvent{
		Event:    "payment.methods_listed",
		TenantID: tenantID.String(),
	})

	return &response.ListPaymentMethodsResponse{
		Items:      items,
		TotalCount: len(items),
	}, nil
}
