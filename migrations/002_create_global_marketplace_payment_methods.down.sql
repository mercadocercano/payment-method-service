-- Rollback de 002_create_global_marketplace_payment_methods.
-- Orden inverso por la FK: primero marketplace (dependiente), luego global.
DROP TABLE IF EXISTS marketplace_payment_methods;
DROP TABLE IF EXISTS global_payment_methods;
