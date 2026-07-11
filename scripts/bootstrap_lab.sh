#!/usr/bin/env bash
# MC-E37 T8 — Bootstrap reproducible de payment_method_db (from-scratch → POS listando métodos).
#
# Orden completo desde cero (los pasos 1-2 son del entorno; este script hace 3-5):
#   1. postgres-setup crea payment_method_db (docker-compose.yml).
#   2. Migraciones 001-004: correr el servicio con DB_USER=postgres UNA vez — el rol de app
#      `payment_method_app` NO puede DDL (T7). Deja schema_migrations version=4 y crea el rol.
#      (En el lab: `docker compose -f docker-compose.yml up -d --no-deps --force-recreate
#       payment-method-service` usa la base sin el override → arranca como postgres.)
#   3. [este script] Fija la password del rol app (004 lo crea SIN password).
#   4. [este script] Seed del catálogo global + pre-seed marketplace.
#   5. [este script] Backfill de las filas tenant (iam_db.tenants ACTIVE × globales activos).
# Después: el servicio vuelve a correr como payment_method_app (docker-compose.override.yml).
#
# Idempotente. Correr desde un entorno con psql que vea las DBs (p.ej. el contenedor lab-postgres):
#   docker cp services/payment-method-service lab-postgres:/tmp/pm-svc
#   docker exec -e APP_PASSWORD=payment_method_app123 lab-postgres bash /tmp/pm-svc/scripts/bootstrap_lab.sh
set -euo pipefail

PAYMENT_DB="${PAYMENT_DB:-payment_method_db}"
APP_ROLE="${APP_ROLE:-payment_method_app}"
APP_PASSWORD="${APP_PASSWORD:?seteá APP_PASSWORD (password dev del rol app, no versionada)}"
export PGUSER="${PGUSER:-postgres}"
HERE="$(cd "$(dirname "$0")" && pwd)"

echo "==> [3/5] password del rol app (out-of-band, no versionada)"
psql -v ON_ERROR_STOP=1 -d "$PAYMENT_DB" -c "ALTER ROLE ${APP_ROLE} PASSWORD '${APP_PASSWORD}';"

echo "==> [4/5] seed global + pre-seed marketplace"
psql -v ON_ERROR_STOP=1 -d "$PAYMENT_DB" -f "${HERE}/../seeds/seed_global_payment_methods.sql"

echo "==> [5/5] backfill filas tenant"
bash "${HERE}/backfill_tenant_payment_methods.sh"

echo "OK — ${PAYMENT_DB} lista (global + marketplace + backfill); el servicio ya puede correr como ${APP_ROLE}."
