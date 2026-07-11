-- Seed: seed_global_payment_methods
-- Description: MC-E37 T2 — Catálogo GLOBAL (control-plane) + pre-seed del plano MARKETPLACE.
--   Fuente ÚNICA de verdad del set global: correr este seed sobre una DB recién migrada (001+002)
--   reproduce el estado del lab (requisito T8: from-scratch == lab).
--
-- SET CANÓNICO = 9 métodos (decisión owner 2026-07-10). Cambio vs el estado previo:
--   - Se quitó el `transfer` suelto (00000001-…005) por redundante con `bank_transfer` (dedup).
--   - Se conserva `qr`. (Antes: 10 filas en el lab / 8 en el seed viejo — ambos divergentes.)
--
-- UUIDs: se PRESERVAN los UUIDs reales que referencia `order_db.pos_sales`
--   (cash = 00000001-…001, mercadopago = 00000001-…006) — decisión (b) del gate T1: SIN remapeo,
--   por eso T4 (pos_sales huérfanos) da 0. Conviven dos familias de UUID (00000001-… y b0000000-…):
--   es intencional; refleja los ids históricos de los métodos, no un error.
--
-- NOTE: seeds NO mutables; se recrean si se borra la DB. Control-plane: el rol de app NO escribe
--   sobre global/marketplace (grants = T7). El plano tenant (`payment_methods`) se puebla por
--   backfill (T3) / onboarding (MC-E38), no acá.

-- ============================================================
-- Plano GLOBAL — catálogo maestro (9 métodos)
-- ============================================================
INSERT INTO global_payment_methods (id, code, name, description, is_active) VALUES
  ('00000001-0000-0000-0000-000000000001', 'cash',          'Efectivo',               'Pago en efectivo al momento de la compra',                          true),
  ('00000001-0000-0000-0000-000000000002', 'debit_card',    'Tarjeta de Débito',      'Pago con tarjeta de débito (Visa Débito, Maestro, etc.)',           true),
  ('00000001-0000-0000-0000-000000000003', 'credit_card',   'Tarjeta de Crédito',     'Pago con tarjeta de crédito (Visa, Mastercard, Amex, Cabal, etc.)', true),
  ('00000001-0000-0000-0000-000000000004', 'qr',            'QR',                     'Pago con código QR interoperable',                                  true),
  ('00000001-0000-0000-0000-000000000006', 'mercadopago',   'Mercado Pago',           'Pago con Mercado Pago (QR, link, cuenta)',                          true),
  ('b0000000-0000-0000-0000-000000000004', 'bank_transfer', 'Transferencia Bancaria', 'Transferencia bancaria CBU/CVU/Alias',                              true),
  ('b0000000-0000-0000-0000-000000000006', 'crypto',        'Criptomonedas',          'Pago con criptomonedas (Bitcoin, USDT, DAI, etc.)',                 true),
  ('b0000000-0000-0000-0000-000000000007', 'on_account',    'Cuenta Corriente',       'Pago a cuenta / cuenta corriente del cliente',                      true),
  ('b0000000-0000-0000-0000-000000000008', 'check',         'Cheque',                 'Pago con cheque (al día o diferido)',                               true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Plano MARKETPLACE — pre-sembrado con el set global ACTIVO (decisión (c) del gate T1)
-- Data-driven: una fila marketplace por cada global activo → siempre en sync con el set anterior.
-- ============================================================
INSERT INTO marketplace_payment_methods (global_payment_method_id, is_active)
SELECT id, true FROM global_payment_methods WHERE is_active = true
ON CONFLICT (global_payment_method_id) DO NOTHING;

DO $$
DECLARE g INT; m INT;
BEGIN
    SELECT count(*) INTO g FROM global_payment_methods;
    SELECT count(*) INTO m FROM marketplace_payment_methods;
    RAISE NOTICE 'Seed MC-E37 completado: global=% | marketplace=% (esperado 9 / 9)', g, m;
END $$;
