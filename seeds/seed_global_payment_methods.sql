-- Seed: seed_global_payment_methods
-- Description: Métodos de pago globales para todos los tenants
-- Author: System
-- Date: 2025-02-09
-- NOTE: Los seeds NO son mutables. Se recrean si se borra la DB.
-- NOTE: No hay endpoints admin para modificarlos en esta fase.

-- ========================================
-- MÉTODOS DE PAGO TRADICIONALES
-- ========================================

-- Método 1: Efectivo (cash)
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000001'::uuid,
    NULL,  -- Global
    'cash',
    'Efectivo',
    'Pago en efectivo al momento de la compra',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- Método 2: Tarjeta de Débito
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000002'::uuid,
    NULL,  -- Global
    'debit_card',
    'Tarjeta de Débito',
    'Pago con tarjeta de débito (Visa Débito, Maestro, etc.)',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- Método 3: Tarjeta de Crédito
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000003'::uuid,
    NULL,  -- Global
    'credit_card',
    'Tarjeta de Crédito',
    'Pago con tarjeta de crédito (Visa, Mastercard, Amex, Cabal, etc.)',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- ========================================
-- TRANSFERENCIAS Y BILLETERAS DIGITALES
-- ========================================

-- Método 4: Transferencia Bancaria
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000004'::uuid,
    NULL,  -- Global
    'bank_transfer',
    'Transferencia Bancaria',
    'Transferencia bancaria CBU/CVU/Alias',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- Método 5: Mercado Pago
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000005'::uuid,
    NULL,  -- Global
    'mercadopago',
    'Mercado Pago',
    'Pago con Mercado Pago (QR, link, cuenta)',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- ========================================
-- CRIPTOMONEDAS Y ALTERNATIVAS
-- ========================================

-- Método 6: Criptomonedas (Crypto)
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000006'::uuid,
    NULL,  -- Global
    'crypto',
    'Criptomonedas',
    'Pago con criptomonedas (Bitcoin, USDT, DAI, etc.)',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- ========================================
-- CRÉDITO Y CUENTA CORRIENTE
-- ========================================

-- Método 7: Cuenta Corriente / A Cuenta
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000007'::uuid,
    NULL,  -- Global
    'on_account',
    'Cuenta Corriente',
    'Pago a cuenta / cuenta corriente del cliente',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- ========================================
-- OTROS MÉTODOS ARGENTINOS
-- ========================================

-- Método 8: Cheque
INSERT INTO payment_methods (id, tenant_id, code, name, description, is_active, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000008'::uuid,
    NULL,  -- Global
    'check',
    'Cheque',
    'Pago con cheque (al día o diferido)',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)) DO NOTHING;

-- Logging
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Payment method seeds completed successfully';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Total global payment methods inserted: 8';
    RAISE NOTICE '';
    RAISE NOTICE 'Métodos disponibles:';
    RAISE NOTICE '  1. Efectivo (cash)';
    RAISE NOTICE '  2. Tarjeta de Débito (debit_card)';
    RAISE NOTICE '  3. Tarjeta de Crédito (credit_card)';
    RAISE NOTICE '  4. Transferencia Bancaria (bank_transfer)';
    RAISE NOTICE '  5. Mercado Pago (mercadopago)';
    RAISE NOTICE '  6. Criptomonedas (crypto)';
    RAISE NOTICE '  7. Cuenta Corriente (on_account)';
    RAISE NOTICE '  8. Cheque (check)';
    RAISE NOTICE '';
    RAISE NOTICE 'NOTE: These seeds are NOT mutable and will be recreated if DB is dropped';
    RAISE NOTICE 'All payment methods are global (tenant_id = NULL) and available for all tenants';
END $$;
