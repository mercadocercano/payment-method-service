-- Migration: 001_create_payment_methods_table
-- Description: Crea la tabla de métodos de pago para el sistema POS (read-only MVP)
-- Author: System
-- Date: 2025-02-09

-- Crear tabla payment_methods
CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,  -- NULL = global, NOT NULL = específico de tenant
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Crear índices para búsqueda eficiente
CREATE INDEX IF NOT EXISTS idx_payment_methods_tenant_id ON payment_methods(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_code_tenant ON payment_methods(code, tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_is_active ON payment_methods(is_active);

-- Constraint para evitar códigos duplicados por tenant
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_methods_unique_code 
    ON payment_methods(code, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid));

-- Comentarios de tabla y columnas
COMMENT ON TABLE payment_methods IS 'Tabla de métodos de pago del sistema (read-only para MVP POS)';
COMMENT ON COLUMN payment_methods.id IS 'Identificador único del método de pago';
COMMENT ON COLUMN payment_methods.tenant_id IS 'ID del tenant (NULL = global, NOT NULL = específico)';
COMMENT ON COLUMN payment_methods.code IS 'Código del método de pago (cash, debit_card, etc.)';
COMMENT ON COLUMN payment_methods.name IS 'Nombre descriptivo del método de pago';
COMMENT ON COLUMN payment_methods.description IS 'Descripción opcional del método de pago';
COMMENT ON COLUMN payment_methods.is_active IS 'Indica si el método de pago está activo';
COMMENT ON COLUMN payment_methods.created_at IS 'Fecha de creación del registro';
COMMENT ON COLUMN payment_methods.updated_at IS 'Fecha de última actualización';
