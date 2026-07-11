-- Migration: 002_create_global_marketplace_payment_methods
-- Description: MC-E37 T2 — Planos GLOBAL y MARKETPLACE de métodos de pago (modelo tres planos con FK).
--   Crea el catálogo maestro (global_payment_methods) y el plano de checkout del marketplace
--   (marketplace_payment_methods, FK al global). NO toca `payment_methods` (plano tenant): ese
--   remodelado (tenant_id NOT NULL + FK a global) es T5. Los datos viven en el seed control-plane
--   (seeds/seed_global_payment_methods.sql), no en la migración, para que "from-scratch" sea
--   reproducible sobre una DB recién migrada (T8).
-- Author: MC-E37 (payment methods tres planos) — sign-off L4 owner 2026-07-10
-- Date: 2026-07-10

-- Plano GLOBAL: catálogo maestro POS (control-plane). Fuente de verdad de code/name.
-- `id` SIN default a propósito: los UUIDs se preservan desde el seed porque
-- `order_db.pos_sales.payment_method_id` los referencia (decisión (b) del gate T1: sin remapeo).
CREATE TABLE IF NOT EXISTS global_payment_methods (
    id UUID PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Plano MARKETPLACE (checkout mc_consumer): mapping de habilitación con FK al global, NO copia.
-- ON DELETE RESTRICT: un global con filas marketplace no se puede borrar; retiro vía is_active=false.
CREATE TABLE IF NOT EXISTS marketplace_payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    global_payment_method_id UUID NOT NULL REFERENCES global_payment_methods(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (global_payment_method_id)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_pm_global ON marketplace_payment_methods(global_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_global_pm_is_active ON global_payment_methods(is_active);

COMMENT ON TABLE global_payment_methods IS 'Catálogo maestro de métodos de pago (control-plane). Fuente de verdad de code/name. PK preservada = UUID que referencia pos_sales (sin remapeo). MC-E37.';
COMMENT ON TABLE marketplace_payment_methods IS 'Plano marketplace (checkout mc_consumer): mapping de habilitación con FK al global. Pre-sembrado con el set global activo. MC-E37 T2.';
