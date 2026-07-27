-- 005_rls_payment_methods.up.sql — Aislamiento fail-closed (Devy RULE-09/RULE-10, roadmap PLAT-E26)
--
-- Retrofit RLS del plano TENANT `payment_methods` (mapping tenant-only tras MC-E37: tenant_id NOT NULL,
-- FK a global_payment_methods). Patrón idéntico a PLAT-E24 (ledger_entries) y PLAT-E25 (sales-service):
-- policy `tenant_isolation` única, sin caso NULL, SIN `break_glass` (escape cross-tenant net-new que el
-- borrador había agregado — eliminado en el gate L4 @dev-architect + @dev-security + owner, 2026-07-11).
--
-- El aislamiento NO depende del WHERE de PostgresPaymentMethodRepository: si una query olvidara filtrar
-- por tenant_id, la base la filtra igual. La app hace SET LOCAL app.tenant_id por transacción
-- (go-shared postgres.WithRLSInTransaction), leído siempre del JWT ya validado por IAM — nunca de un
-- header/param crudo.
--
-- Planos control-plane (global_payment_methods, marketplace_payment_methods) NO llevan RLS por tenant:
-- son catálogos compartidos (todo tenant ve el mismo catálogo global — una policy tenant_id = ... sería
-- semánticamente incorrecta). Su fail-closed lo garantizan los grants de MC-E37 004 (rol
-- payment_method_app con SELECT-only en ambos, sin INSERT/UPDATE/DELETE), no RLS. Ver PLAT-E26 §T2.

ALTER TABLE payment_methods ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_methods FORCE  ROW LEVEL SECURITY;   -- aplica incluso al dueño de la tabla

DROP POLICY IF EXISTS tenant_isolation ON payment_methods;
CREATE POLICY tenant_isolation ON payment_methods
    USING      (tenant_id = current_setting('app.tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);
