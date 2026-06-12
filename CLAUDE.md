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

## Memoria persistente (Engram)

Tenés acceso a memoria persistente entre sesiones vía las herramientas MCP de Engram (`mem_save`, `mem_search`, `mem_context`, etc.). Proyecto: **`mercado-cercano`** (memoria unificada del ecosistema; este service es un polyrepo con su `.git` propio, pero comparte memoria con el resto vía `.engram/config.json`).

**Cuándo guardar** — sin esperar que te lo pidan:
- Al resolver un bug no trivial: síntoma, causa raíz, fix aplicado.
- Al tomar una decisión de diseño: qué se decidió y por qué.
- Al descubrir un patrón o convención del proyecto que no está documentada.
- Al completar una feature o refactor significativo: qué cambió y dónde.

**Cuándo buscar** — antes de empezar cualquier tarea:
- `mem_context` al inicio de sesión o tras una compaction para recuperar el estado anterior.
- `mem_search` cuando el usuario menciona algo que puede tener historial ("el bug de autenticación", "la migración de la semana pasada").

**Al cerrar sesión**: llamar `mem_session_summary` para dejar un resumen recuperable.
