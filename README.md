# payment-method-service

Servicio de gestión de métodos de pago para el ecosistema SaaS Multi-Tenant "Mercado Cercano".

## 🎯 Responsabilidad

**Proporcionar métodos de pago disponibles (read-only) para el flujo POS sin procesamiento de pagos**.

## 📋 Alcance Actual (MVP POS)

### ✅ Qué HACE
- ✅ Listar métodos de pago disponibles para un tenant
- ✅ Retornar métodos globales + específicos del tenant
- ✅ Filtrar por activos/inactivos
- ✅ Obtener detalle de un método de pago por ID
- ✅ Aislamiento multi-tenant estricto

### ❌ Qué NO hace (fuera de alcance)
- ❌ Procesar pagos reales (integración con pasarelas)
- ❌ Validar tarjetas o medios de pago
- ❌ Calcular comisiones o recargos
- ❌ Gestionar cuotas o financiación
- ❌ Conciliación bancaria
- ❌ Configuración de terminales POS físicas
- ❌ Webhooks de notificaciones de pago
- ❌ Crear/Actualizar/Eliminar métodos (admin vendrá después)
- ❌ Integración con Mercado Pago, Stripe, etc.

## 🏗️ Arquitectura

- **Patrón**: Arquitectura Hexagonal + DDD
- **Framework**: Go + Gin
- **Base de datos**: PostgreSQL
- **Puerto interno**: 8080
- **Puerto externo**: 8140

## 📡 API Endpoints

### GET /api/v1/payment-methods
Lista métodos de pago disponibles para el tenant.

**Headers**:
- `X-Tenant-ID`: UUID del tenant (requerido)

**Query Parameters**:
- `active_only`: `true` o `false` (default: `true`)

**Response**:
```json
{
  "items": [
    {
      "id": "uuid",
      "tenant_id": null,
      "code": "cash",
      "name": "Efectivo",
      "description": "Pago en efectivo al momento de la compra",
      "is_active": true,
      "is_global": true,
      "created_at": "2025-02-09T10:00:00Z",
      "updated_at": "2025-02-09T10:00:00Z"
    },
    {
      "id": "uuid",
      "tenant_id": null,
      "code": "debit_card",
      "name": "Tarjeta de Débito",
      "description": "Pago con tarjeta de débito",
      "is_active": true,
      "is_global": true,
      "created_at": "2025-02-09T10:00:00Z",
      "updated_at": "2025-02-09T10:00:00Z"
    }
  ],
  "total_count": 5
}
```

### GET /api/v1/payment-methods/:id
Obtiene detalle de un método de pago específico.

**Headers**:
- `X-Tenant-ID`: UUID del tenant (requerido)

**Path Parameters**:
- `id`: UUID del método de pago

**Response**:
```json
{
  "id": "uuid",
  "tenant_id": null,
  "code": "cash",
  "name": "Efectivo",
  "description": "Pago en efectivo al momento de la compra",
  "is_active": true,
  "is_global": true,
  "created_at": "2025-02-09T10:00:00Z",
  "updated_at": "2025-02-09T10:00:00Z"
}
```

## 🗄️ Modelo de Datos

```sql
CREATE TABLE payment_methods (
    id UUID PRIMARY KEY,
    tenant_id UUID,           -- NULL = global, NOT NULL = específico
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

**Lógica de consulta**:
```sql
-- Retorna métodos globales (tenant_id IS NULL) + métodos del tenant
SELECT * FROM payment_methods 
WHERE (tenant_id IS NULL OR tenant_id = :tenant_id)
AND is_active = true
ORDER BY (tenant_id IS NULL) DESC, name ASC
```

**Índices**:
- `idx_payment_methods_tenant_id` en `(tenant_id)`
- `idx_payment_methods_code_tenant` en `(code, tenant_id)`
- `idx_payment_methods_is_active` en `(is_active)`
- `idx_payment_methods_unique_code` UNIQUE en `(code, COALESCE(tenant_id, '00000000...'))`

## 🌱 Seeds

Los seeds **NO son mutables**. Se recrean si se borra la DB.

**Métodos de pago globales** (tenant_id = NULL):
- ✅ Efectivo (`cash`)
- ✅ Tarjeta de Débito (`debit_card`)
- ✅ Tarjeta de Crédito (`credit_card`)
- ✅ Transferencia Bancaria (`transfer`)
- ✅ Mercado Pago (`mercadopago`)

Todos disponibles para todos los tenants sin configuración adicional.

## 🚀 Desarrollo

### Prerrequisitos
- Go 1.22+
- PostgreSQL 15+
- Docker (opcional)

### Instalación Local

```bash
# Instalar dependencias
go mod download

# Ejecutar migraciones
psql -h localhost -U postgres -d payment_method_db -f migrations/001_create_payment_methods_table.sql

# Ejecutar seeds
psql -h localhost -U postgres -d payment_method_db -f seeds/seed_global_payment_methods.sql

# Ejecutar servicio
go run src/main.go
```

### Docker

```bash
# Build imagen
docker build -t payment-method-service .

# Ejecutar contenedor
docker run -p 8140:8080 \
  -e DB_HOST=postgres \
  -e DB_NAME=payment_method_db \
  payment-method-service
```

### Docker Compose (Desarrollo)

```bash
# Desde la raíz del monorepo
make lite-start

# El servicio estará disponible en:
# http://localhost:8140/health
```

## 🧪 Testing

```bash
# Tests unitarios
go test ./...

# Test manual del endpoint
curl -H "X-Tenant-ID: <tenant-uuid>" \
  http://localhost:8140/api/v1/payment-methods
```

## 📦 Estructura del Proyecto

```
payment-method-service/
├── src/
│   ├── payment_method/
│   │   ├── domain/
│   │   │   ├── entity/          # PaymentMethod entity
│   │   │   └── port/            # PaymentMethodRepository interface
│   │   ├── application/
│   │   │   ├── usecase/         # GetByID, List
│   │   │   └── response/        # DTOs
│   │   └── infrastructure/
│   │       ├── controller/      # HTTP handlers
│   │       ├── persistence/     # PostgreSQL repository
│   │       └── config/          # Module setup
│   ├── shared/                  # Shared utilities
│   └── main.go                  # Entry point
├── migrations/                  # SQL migrations
├── seeds/                       # Initial data
├── Dockerfile                   # Production
├── Dockerfile.dev               # Development
└── README.md
```

## 🔐 Multi-Tenancy

- **Mecanismo único**: Header `X-Tenant-ID`
- **Métodos globales**: `tenant_id = NULL` (disponibles para todos)
- **Métodos específicos**: `tenant_id != NULL` (solo para ese tenant)
- **Query lógica**: `WHERE (tenant_id IS NULL OR tenant_id = :tenant_id)`

## 📊 Monitoreo

- **Health Check**: `GET /health`
- **Prometheus Metrics**: `GET /metrics` (si `PROMETHEUS_ENABLED=true`)

## 🔗 Integración con POS

El servicio se integra con el flujo POS para:
1. Listar métodos de pago disponibles
2. Seleccionar método en la venta
3. Incluir `payment_method_id` en el payload de orden
4. Mostrar método de pago en reportes

## 📝 Migración desde Monorepo

Este servicio fue creado como parte del proyecto SaaS Multi-Tenant.

**Repositorio**: https://github.com/mercadocercano/payment-method-service

## 🤝 Contribución

1. Fork el proyecto
2. Crea una rama para tu feature
3. Commit tus cambios
4. Push a la rama
5. Abre un Pull Request

## 📄 Licencia

MIT License
