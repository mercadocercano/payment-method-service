# CLAUDE.md — payment-method-service

Métodos de pago **solo lectura** para POS: globales (`tenant_id` NULL) + filas por tenant.

**Puerto**: 8080 (contenedor); **8140** en mapeo host típico | **Stack**: Go + Gin + PostgreSQL

Habla siempre en español.

## Comandos

```bash
go run src/main.go
go test ./...
```

## Contexto on-demand

| Archivo | Uso |
|---------|-----|
| `payment-method-service-management/api-endpoints.md` | GET listado/detalle, JSON |
| `payment-method-service-management/architecture.md` | Hexagonal, SQL, seeds |
| `payment-method-service-management/config.md` | Env, Docker, K8s/CI |

## Reglas compartidas

`ai-tools/rules/architecture.md`, `multi-tenant.md`, `api-gateway.md`, `api-response-format.md`.

Cabecera **`X-Tenant-ID`** (UUID) en rutas de API.
