-- 004_create_app_role.up.sql — MC-E37 T7 (RULE-09: rol de aplicación sin DDL ni escritura de control-plane)
--
-- El runtime de payment-method-service conectaba como `postgres` (superuser) — el mismo default
-- inseguro que sales-service tenía antes de PLAT-E25 (ver sales-service/migrations/025). Este rol
-- least-privilege lo reemplaza en runtime:
--   • LEE los tres planos (el POS lee payment_methods con JOIN a global).
--   • ESCRIBE solo el plano tenant (payment_methods) — provisión MC-E38 / toggle MC-E39 futuros.
--   • global_payment_methods y marketplace_payment_methods son CONTROL-PLANE → el rol de app NO
--     escribe (T7 DoD: `INSERT INTO global_payment_methods …` con este rol → permission denied).
-- NOBYPASSRLS: forward-compatible con el retrofit RLS de PLAT-E26.
--
-- Idempotente (CREATE ROLE guardado en IF NOT EXISTS; GRANT es idempotente en Postgres).
--
-- Seguridad (mismo patrón que sales_app/025): esta migración NO fija password — un literal quedaría
-- en el git history para siempre al commitear. El rol se crea con LOGIN pero SIN password (login
-- deshabilitado hasta setearla); la password real se fija out-of-band vía
-- `ALTER ROLE payment_method_app PASSWORD '...'` corrido a mano contra lab-postgres, nunca versionado.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'payment_method_app') THEN
    CREATE ROLE payment_method_app LOGIN
      NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
  END IF;
END
$$;

GRANT CONNECT ON DATABASE payment_method_db TO payment_method_app;
GRANT USAGE ON SCHEMA public TO payment_method_app;

-- Lectura de los tres planos.
GRANT SELECT ON global_payment_methods       TO payment_method_app;
GRANT SELECT ON marketplace_payment_methods  TO payment_method_app;
GRANT SELECT ON payment_methods              TO payment_method_app;

-- Escritura SOLO en el plano tenant (control-plane global/marketplace queda como postgres).
GRANT INSERT, UPDATE, DELETE ON payment_methods TO payment_method_app;

-- Runtime solo chequea la versión al arrancar (no aplica migraciones nuevas → SELECT alcanza).
GRANT SELECT ON schema_migrations TO payment_method_app;
