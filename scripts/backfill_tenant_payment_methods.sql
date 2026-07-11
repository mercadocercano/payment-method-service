-- Backfill: filas de habilitación por tenant (plano tenant de payment_methods) — MC-E37 T3.
-- Inserta una fila por cada (tenant × método global ACTIVO). Idempotente (ON CONFLICT DO NOTHING).
--
-- FUENTE DE TENANTS (corrección del gate T1, sign-off owner 2026-07-10): el registro canónico es
-- `iam_db.tenants` (29 ACTIVE), NO `tenant_db.tenants` (que no existe — tenant_db solo tiene
-- points_of_sale / tenant_config / tenant_settings). Alcance = TODOS los tenants ACTIVE (decisión (c)).
--
-- CROSS-DB: `iam_db.tenants` y `payment_method_db.payment_methods` viven en DBs distintas de la misma
-- instancia → NO se puede en una sola conexión SQL. Este .sql es el paso 2 (destino); el paso 1
-- (lista de tenants) lo inyecta el wrapper `backfill_tenant_payment_methods.sh`, que sustituye
-- :'tenant_values' por la lista `('uuid'),('uuid'),...` leída de iam_db. Ejecutar preferentemente
-- vía el .sh; este archivo documenta/versiona el INSERT canónico.
--
-- Uso directo (si ya tenés la lista):
--   psql -d payment_method_db -v tenant_values="('id1'),('id2')" -f backfill_tenant_payment_methods.sql

INSERT INTO payment_methods (tenant_id, global_payment_method_id, is_active)
SELECT t.tenant_id::uuid, g.id, true
FROM (VALUES :tenant_values) AS t(tenant_id)
CROSS JOIN global_payment_methods g
WHERE g.is_active = true
ON CONFLICT (tenant_id, global_payment_method_id) DO NOTHING;
