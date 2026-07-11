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
		SELECT pm.id, pm.tenant_id, g.code, g.name, g.description, pm.is_active, pm.created_at, pm.updated_at
		FROM payment_methods pm
		JOIN global_payment_methods g ON g.id = pm.global_payment_method_id
		WHERE pm.id = $1 AND pm.tenant_id = $2
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

// FindAll retorna los métodos del plano tenant, con code/name resueltos del catálogo global.
func (r *PostgresPaymentMethodRepository) FindAll(tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error) {
	// Plano tenant (mapping) + JOIN al catálogo global para traer code/name (MC-E37 T6).
	query := `
		SELECT pm.id, pm.tenant_id, g.code, g.name, g.description, pm.is_active, pm.created_at, pm.updated_at
		FROM payment_methods pm
		JOIN global_payment_methods g ON g.id = pm.global_payment_method_id
		WHERE pm.tenant_id = $1
	`

	args := []interface{}{tenantID}

	// Disponible = habilitado por el tenant Y activo a nivel global (un método retirado a nivel
	// global desaparece para todos los tenants, aunque lo tuvieran habilitado).
	if activeOnly {
		query += ` AND pm.is_active = true AND g.is_active = true`
	}

	// Orden por nombre del catálogo global.
	query += ` ORDER BY g.name ASC`

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
