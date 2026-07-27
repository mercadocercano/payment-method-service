#!/usr/bin/env bash
# MC-E37 T3 — Backfill de filas de habilitación por tenant (plano tenant de payment_methods).
#
# Inserta una fila por cada (tenant ACTIVE × método global ACTIVO). Idempotente.
# CROSS-DB: lee la lista de tenants del registro canónico `iam_db.tenants` (corrección del gate T1,
# sign-off owner 2026-07-10 — NO `tenant_db.tenants`, que no existe) e inserta en
# `payment_method_db.payment_methods`. Ambas DBs, misma instancia postgres.
#
# Requiere psql con acceso a las dos DBs (correr dentro del contenedor postgres o con PG* apuntando
# a la instancia). Variables (con defaults del lab):
#   IAM_DB=iam_db  PAYMENT_DB=payment_method_db  PGUSER=postgres
set -euo pipefail

IAM_DB="${IAM_DB:-iam_db}"
PAYMENT_DB="${PAYMENT_DB:-payment_method_db}"
export PGUSER="${PGUSER:-postgres}"

# Paso 1 — lista de tenants ACTIVE del registro canónico.
TENANT_IDS="$(psql -tA -v ON_ERROR_STOP=1 -d "$IAM_DB" -c "SELECT id FROM tenants WHERE status='ACTIVE';")"
[ -n "$TENANT_IDS" ] || { echo "ERROR: 0 tenants ACTIVE en ${IAM_DB}.tenants" >&2; exit 1; }
N_TENANTS="$(printf '%s\n' "$TENANT_IDS" | grep -c .)"

# Construir la lista VALUES ('uuid'),('uuid'),...
TENANT_VALUES="$(printf '%s\n' "$TENANT_IDS" | awk 'NF{printf "(\047%s\047),", $0}' | sed 's/,$//')"

# Paso 2 — backfill idempotente (tenant × global activo).
psql -v ON_ERROR_STOP=1 -d "$PAYMENT_DB" \
  -v tenant_values="$TENANT_VALUES" \
  -f "$(dirname "$0")/backfill_tenant_payment_methods.sql"

FINAL="$(psql -tA -d "$PAYMENT_DB" -c "SELECT count(*) FROM payment_methods;")"
echo "Backfill MC-E37 T3 OK: ${N_TENANTS} tenants ACTIVE × globales activos → payment_methods=${FINAL} filas"
