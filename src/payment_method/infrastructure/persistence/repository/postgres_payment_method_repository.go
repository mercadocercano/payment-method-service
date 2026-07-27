package repository

import (
	"context"
	"database/sql"
	"fmt"
	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/domain/port"

	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/postgres"
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

// FindByID busca un método de pago habilitado del tenant por su ID.
// La query corre dentro de WithRLSInTransaction: la policy RLS `tenant_isolation` (E26/E24)
// exige `app.tenant_id` seteado en la transacción para evaluar cualquier lectura de
// payment_methods. El filtro manual `WHERE pm.tenant_id = $2` se mantiene como defensa en
// profundidad (la policy ya lo garantiza).
func (r *PostgresPaymentMethodRepository) FindByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*entity.PaymentMethod, error) {
	rc := postgres.RLSContext{TenantID: tenantID.String()}

	var pm entity.PaymentMethod
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		query := `
			SELECT pm.id, pm.tenant_id, g.code, g.name, g.description, pm.is_active, pm.created_at, pm.updated_at
			FROM payment_methods pm
			JOIN global_payment_methods g ON g.id = pm.global_payment_method_id
			WHERE pm.id = $1 AND pm.tenant_id = $2
		`
		return tx.QueryRowContext(ctx, query, id, tenantID).Scan(
			&pm.ID,
			&pm.TenantID,
			&pm.Code,
			&pm.Name,
			&pm.Description,
			&pm.IsActive,
			&pm.CreatedAt,
			&pm.UpdatedAt,
		)
	})

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error finding payment method: %w", err)
	}

	return &pm, nil
}

// FindAll retorna los métodos del plano tenant, con code/name resueltos del catálogo global.
// Corre dentro de WithRLSInTransaction por la policy RLS `tenant_isolation` (ver FindByID).
// El JOIN a global_payment_methods (control-plane, sin RLS por tenant) no se ve afectado por
// `SET LOCAL app.tenant_id`; solo se filtran las filas de payment_methods.
func (r *PostgresPaymentMethodRepository) FindAll(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*entity.PaymentMethod, error) {
	rc := postgres.RLSContext{TenantID: tenantID.String()}

	paymentMethods := make([]*entity.PaymentMethod, 0)
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
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

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("error querying payment methods: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pm entity.PaymentMethod
			if err := rows.Scan(
				&pm.ID,
				&pm.TenantID,
				&pm.Code,
				&pm.Name,
				&pm.Description,
				&pm.IsActive,
				&pm.CreatedAt,
				&pm.UpdatedAt,
			); err != nil {
				return fmt.Errorf("error scanning payment method: %w", err)
			}
			paymentMethods = append(paymentMethods, &pm)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating payment methods: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paymentMethods, nil
}
