package repository

import (
	"database/sql"
	"fmt"
	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/domain/port"

	"github.com/google/uuid"
)

// PostgresPaymentMethodRepository implementa PaymentMethodRepository usando PostgreSQL
type PostgresPaymentMethodRepository struct {
	db *sql.DB
}

// NewPostgresPaymentMethodRepository crea una nueva instancia del repositorio
func NewPostgresPaymentMethodRepository(db *sql.DB) port.PaymentMethodRepository {
	return &PostgresPaymentMethodRepository{
		db: db,
	}
}

// FindByID busca un método de pago por su ID (global o del tenant)
func (r *PostgresPaymentMethodRepository) FindByID(id uuid.UUID, tenantID uuid.UUID) (*entity.PaymentMethod, error) {
	query := `
		SELECT id, tenant_id, code, name, description, is_active, created_at, updated_at
		FROM payment_methods
		WHERE id = $1 AND (tenant_id IS NULL OR tenant_id = $2)
	`

	var pm entity.PaymentMethod
	err := r.db.QueryRow(query, id, tenantID).Scan(
		&pm.ID,
		&pm.TenantID,
		&pm.Code,
		&pm.Name,
		&pm.Description,
		&pm.IsActive,
		&pm.CreatedAt,
		&pm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error finding payment method: %w", err)
	}

	return &pm, nil
}

// FindAll retorna todos los métodos de pago disponibles para un tenant
// (incluye métodos globales + específicos del tenant)
func (r *PostgresPaymentMethodRepository) FindAll(tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error) {
	// Query base: métodos globales + métodos del tenant
	query := `
		SELECT id, tenant_id, code, name, description, is_active, created_at, updated_at
		FROM payment_methods
		WHERE (tenant_id IS NULL OR tenant_id = $1)
	`

	args := []interface{}{tenantID}

	// Filtrar solo activos si se solicita
	if activeOnly {
		query += ` AND is_active = true`
	}

	// Ordenar por: globales primero, luego por nombre
	query += ` ORDER BY (tenant_id IS NULL) DESC, name ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying payment methods: %w", err)
	}
	defer rows.Close()

	paymentMethods := make([]*entity.PaymentMethod, 0)
	for rows.Next() {
		var pm entity.PaymentMethod
		err := rows.Scan(
			&pm.ID,
			&pm.TenantID,
			&pm.Code,
			&pm.Name,
			&pm.Description,
			&pm.IsActive,
			&pm.CreatedAt,
			&pm.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning payment method: %w", err)
		}
		paymentMethods = append(paymentMethods, &pm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment methods: %w", err)
	}

	return paymentMethods, nil
}
