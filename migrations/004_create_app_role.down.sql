-- Rollback de 004_create_app_role (best-effort).
REVOKE ALL ON schema_migrations             FROM payment_method_app;
REVOKE ALL ON payment_methods               FROM payment_method_app;
REVOKE ALL ON marketplace_payment_methods   FROM payment_method_app;
REVOKE ALL ON global_payment_methods        FROM payment_method_app;
REVOKE USAGE ON SCHEMA public               FROM payment_method_app;
REVOKE CONNECT ON DATABASE payment_method_db FROM payment_method_app;
DROP ROLE IF EXISTS payment_method_app;
