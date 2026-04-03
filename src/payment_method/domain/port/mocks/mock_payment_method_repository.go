package mocks

import (
	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/domain/port"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPaymentMethodRepository es un mock del repositorio para tests
type MockPaymentMethodRepository struct {
	mock.Mock
}

// Verificar que implementa la interfaz en tiempo de compilación
var _ port.PaymentMethodRepository = (*MockPaymentMethodRepository)(nil)

func (m *MockPaymentMethodRepository) FindByID(id uuid.UUID, tenantID uuid.UUID) (*entity.PaymentMethod, error) {
	args := m.Called(id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PaymentMethod), args.Error(1)
}

func (m *MockPaymentMethodRepository) FindAll(tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error) {
	args := m.Called(tenantID, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PaymentMethod), args.Error(1)
}
