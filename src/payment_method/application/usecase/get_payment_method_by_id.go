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
	logger     port.PaymentEventLogger
}

// NewGetPaymentMethodByIDUseCase crea una nueva instancia del caso de uso
func NewGetPaymentMethodByIDUseCase(repository port.PaymentMethodRepository, logger port.PaymentEventLogger) *GetPaymentMethodByIDUseCase {
	return &GetPaymentMethodByIDUseCase{
		repository: repository,
		logger:     logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (uc *GetPaymentMethodByIDUseCase) logEvent(e port.PaymentEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
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
		uc.logEvent(port.PaymentEvent{
			Event:           "payment.method_fetch_failed",
			TenantID:        tenantID.String(),
			PaymentMethodID: id.String(),
			Reason:          err.Error(),
		})
		return nil, err
	}

	if paymentMethod == nil {
		uc.logEvent(port.PaymentEvent{
			Event:           "payment.method_not_found",
			TenantID:        tenantID.String(),
			PaymentMethodID: id.String(),
		})
		return nil, errors.New("payment method not found")
	}

	uc.logEvent(port.PaymentEvent{
		Event:             "payment.method_found",
		TenantID:          tenantID.String(),
		PaymentMethodID:   paymentMethod.ID.String(),
		PaymentMethodCode: paymentMethod.Code,
	})

	// Convertir a response
	return response.FromEntity(paymentMethod), nil
}
