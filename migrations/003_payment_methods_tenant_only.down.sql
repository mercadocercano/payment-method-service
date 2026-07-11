-- Rollback de 003_payment_methods_tenant_only (best-effort: no restaura datos de code/name).
ALTER TABLE payment_methods DROP CONSTRAINT IF EXISTS uq_payment_methods_tenant_global;
DROP INDEX IF EXISTS idx_payment_methods_global;
ALTER TABLE payment_methods DROP CONSTRAINT IF EXISTS fk_payment_methods_global;
ALTER TABLE payment_methods ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE payment_methods DROP COLUMN IF EXISTS global_payment_method_id;

ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS code VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE payment_methods ALTER COLUMN code DROP DEFAULT;
ALTER TABLE payment_methods ALTER COLUMN name DROP DEFAULT;

CREATE INDEX IF NOT EXISTS idx_payment_methods_code_tenant ON payment_methods(code, tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_methods_unique_code
  ON payment_methods(code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid));
