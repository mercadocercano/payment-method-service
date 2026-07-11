-- Migration: 003_payment_methods_tenant_only
-- Description: MC-E37 T3/T5 (re-secuenciadas, sign-off owner 2026-07-10) — `payment_methods` pasa a
--   PLANO TENANT puro: mapping (tenant_id, global_payment_method_id) con FK al catálogo global e
--   is_active propio (toggle per-tenant). code/name/description viven SOLO en el global (decisión
--   (a) del gate T1: mapping puro, sin override).
--   Se re-secuenció respecto del roadmap original (T3 backfill dependía de T5 remodelado): como T2
--   dejó la tabla VACÍA, el remodelado destructivo es trivial y DEBE ir ANTES del backfill. Esta 003
--   absorbe lo que el roadmap llamaba T5.
--   Requiere que 002 (global_payment_methods) exista.
-- Author: MC-E37 — sign-off L4 owner 2026-07-10
-- Date: 2026-07-10

-- Absorbe old-T5: las filas globales (tenant_id IS NULL) ya viven en global_payment_methods (T2) →
-- se eliminan. En el lab la tabla está vacía (no-op); protege el caso prod con globales NULL viejos.
-- No hay filas tenant (tenant_id NOT NULL) con code/name propios → dropear code/name no pierde datos.
DELETE FROM payment_methods WHERE tenant_id IS NULL;

-- FK al catálogo global (mapping puro). Tabla vacía → se puede exigir NOT NULL directo.
ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS global_payment_method_id UUID;
ALTER TABLE payment_methods
  ADD CONSTRAINT fk_payment_methods_global
  FOREIGN KEY (global_payment_method_id) REFERENCES global_payment_methods(id) ON DELETE RESTRICT;

-- Mapping puro: code/name/description viven SOLO en el catálogo global.
ALTER TABLE payment_methods DROP COLUMN IF EXISTS code;
ALTER TABLE payment_methods DROP COLUMN IF EXISTS name;
ALTER TABLE payment_methods DROP COLUMN IF EXISTS description;

-- Tenant-only: sin filas globales (NULL) → listo para el patrón RLS simple de PLAT-E24 (desbloquea PLAT-E26).
ALTER TABLE payment_methods ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE payment_methods ALTER COLUMN global_payment_method_id SET NOT NULL;

-- Índices: fuera los del esquema viejo (COALESCE + code), dentro el UNIQUE del mapping.
DROP INDEX IF EXISTS idx_payment_methods_unique_code;
DROP INDEX IF EXISTS idx_payment_methods_code_tenant;
ALTER TABLE payment_methods
  ADD CONSTRAINT uq_payment_methods_tenant_global UNIQUE (tenant_id, global_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_global ON payment_methods(global_payment_method_id);

COMMENT ON TABLE payment_methods IS 'Plano TENANT (mapping de habilitación): una fila por (tenant × método global provisionado), FK al global + is_active propio (toggle per-tenant). Provisión por backfill (T3) / onboarding (MC-E38). MC-E37.';
COMMENT ON COLUMN payment_methods.global_payment_method_id IS 'FK al catálogo global (code/name viven allí — mapping puro).';
COMMENT ON COLUMN payment_methods.is_active IS 'Habilitación del método por el tenant (toggle). El POS lo muestra si pm.is_active AND global.is_active.';
