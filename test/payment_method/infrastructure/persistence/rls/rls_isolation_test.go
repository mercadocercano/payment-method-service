//go:build integration

// Test de aislamiento cross-tenant adversarial (PLAT-E26 T6).
//
// Levanta un Postgres efímero (testcontainers), aplica TODAS las migraciones de esquema
// reales del servicio (incluida 005_rls_payment_methods, RLS fail-closed) y conecta bajo un
// rol sin BYPASSRLS (mismo patrón que `payment_method_app` en lab-postgres, creado por MC-E37
// 004, verificado en T3) — un superuser SIEMPRE bypasea RLS aunque la tabla tenga FORCE ROW
// LEVEL SECURITY, así que probar contra él daría falsos verdes (probado empíricamente en E24 y
// E25). El `is_active_integration_test.go` de MC-E37 corre como superuser a propósito (siembra
// el catálogo global de control-plane) y NO sirve para este test: no ejerce la policy.
//
// Cobertura (T6 "Hecho cuando", patrón E24/E25 adaptado al plano único con FK al catálogo):
//   - payment_methods (plano tenant, columna tenant_id propia): A la ve, B NO la ve.
//   - INSERT con tenant_id=A bajo sesión de tenant B rechazado por WITH CHECK.
//   - UPDATE cross-tenant es no-op (la policy USING oculta la fila a B — relevante para el
//     futuro toggle per-tenant de MC-E39, primera escritura per-tenant).
//   - JOIN a global_payment_methods (control-plane, sin RLS) sigue resolviendo code/name bajo
//     RLS activo (Riesgo de la épica: el SET LOCAL app.tenant_id no debe romper el JOIN).
//   - Fail-closed sin contexto de tenant: ninguna query se satura por accidente.
//
// El contenedor se levanta UNA sola vez para todo el binario de test vía TestMain (no por
// TestXxx) — así los ~3-5s de arranque no se pagan por cada función de test.
//
// Para correrlo:
//
//	go test -tags=integration ./test/payment_method/infrastructure/persistence/rls/... -run CrossTenant -v
package rls_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hornosg/go-shared/infrastructure/postgres"

	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/infrastructure/persistence/repository"
)

// containerDB es el nombre de la base del contenedor efímero — el rol de app necesita CONNECT
// sobre ESTA base (004_create_app_role hardcodea `payment_method_db`, el nombre real de
// lab-postgres, que no existe acá; por eso 004 se excluye del apply y el rol se replica abajo).
const containerDB = "payment_test"

// appRoleName/appRolePassword replican en el Postgres efímero el rol `payment_method_app`
// creado en lab-postgres por MC-E37 004: sin este rol, la conexión de test cae al usuario
// `postgres` del contenedor (superuser) y un superuser SIEMPRE bypasea RLS — FORCE ROW LEVEL
// SECURITY no lo alcanza.
const (
	appRoleName     = "payment_method_app_test"
	appRolePassword = "payment_method_app_test"
)

// appDB es la única conexión que usan los TestXxx de este archivo — bajo el rol sin
// BYPASSRLS, la que de verdad queda sujeta a la policy `tenant_isolation` de 005.
var appDB *sql.DB

// globalMethodID/globalCode/globalName identifican el método del catálogo global (control-plane)
// sembrado en TestMain con el superuser: las filas de `payment_methods` lo referencian por FK, y
// las lecturas del repository resuelven code/name desde él vía JOIN.
var (
	globalMethodID uuid.UUID
	globalCode     string
	globalName     = "Test Method (E26 T6)"
)

// TestMain levanta el Postgres efímero, aplica las migraciones de esquema reales, crea el rol
// de aplicación y siembra el catálogo global UNA sola vez para todo el binario — se comparte
// entre todas las TestXxx en vez de pagar el arranque del contenedor por cada una.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(containerDB),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("error starting postgres container: %v", err)
	}

	superConnStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("error getting connection string: %v", err)
	}

	superDB, err := sql.Open("postgres", superConnStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	if err := superDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	if err := applyMigrations(superDB); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}
	if err := createAppRole(superDB); err != nil {
		log.Fatalf("error creating app role: %v", err)
	}
	if err := seedGlobalCatalog(superDB); err != nil {
		log.Fatalf("error seeding global catalog: %v", err)
	}

	appConnStr, err := withCredentials(superConnStr, appRoleName, appRolePassword)
	if err != nil {
		log.Fatalf("error building app connection string: %v", err)
	}
	appDB, err = sql.Open("postgres", appConnStr)
	if err != nil {
		log.Fatalf("error opening app-role database: %v", err)
	}
	if err := appDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database as %s: %v", appRoleName, err)
	}

	code := m.Run()

	_ = appDB.Close()
	_ = superDB.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

// applyMigrations corre en orden TODAS las migraciones .up.sql de ESQUEMA de payment-method-service
// (001..005, incluida 005_rls_payment_methods) — el mismo esquema que corre en
// lab-postgres/payment_method_db tras PLAT-E26 T2. 004_create_app_role queda afuera a propósito:
// es role/grant DDL que hardcodea `GRANT CONNECT ON DATABASE payment_method_db` (el nombre real de
// la DB en lab-postgres) y `GRANT ... ON schema_migrations` (tabla que este apply directo no crea).
// No existe una DB "payment_method_db" en el contenedor efímero (creado como "payment_test"), y
// este archivo ya replica el rol de aplicación para el container vía createAppRole() con sus
// propios grants.
func applyMigrations(db *sql.DB) error {
	dir := "../../../../../migrations"
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range dirEntries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".sql" || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		if name == "004_create_app_role.up.sql" {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no se encontraron migraciones .up.sql en %s", dir)
	}

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("migración %s: %w", f, err)
		}
	}
	return nil
}

// createAppRole crea (con el superuser) un rol sin DDL ni BYPASSRLS y le otorga los mismos
// grants efectivos que 004_create_app_role.up.sql otorga a `payment_method_app` en lab-postgres
// (menos el SELECT sobre schema_migrations, que no existe en este apply directo) — réplica del
// rol real, no un stand-in simplificado: SELECT en los tres planos + INSERT/UPDATE/DELETE solo
// en el plano tenant `payment_methods` (control-plane global/marketplace = SELECT-only).
func createAppRole(superDB *sql.DB) error {
	stmts := []string{
		`CREATE ROLE ` + appRoleName + ` WITH LOGIN PASSWORD '` + appRolePassword + `' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS`,
		`GRANT CONNECT ON DATABASE ` + containerDB + ` TO ` + appRoleName,
		`GRANT USAGE ON SCHEMA public TO ` + appRoleName,
		`GRANT SELECT ON global_payment_methods, marketplace_payment_methods, payment_methods TO ` + appRoleName,
		`GRANT INSERT, UPDATE, DELETE ON payment_methods TO ` + appRoleName,
	}
	for _, stmt := range stmts {
		if _, err := superDB.Exec(stmt); err != nil {
			return fmt.Errorf("stmt %q: %w", stmt, err)
		}
	}
	return nil
}

// seedGlobalCatalog inserta un método en el catálogo global con el superuser (control-plane sin
// RLS; el rol de app tiene SELECT-only ahí). Su `id` sin default se fija a mano — igual que en
// lab-postgres, donde los UUIDs globales se preservan porque `order_db.pos_sales.payment_method_id`
// los referencia. Las filas de `payment_methods` de los tests lo referencian por FK.
func seedGlobalCatalog(superDB *sql.DB) error {
	globalMethodID = uuid.New()
	globalCode = "zzz_e26t6_" + globalMethodID.String()[:8]
	_, err := superDB.Exec(
		`INSERT INTO global_payment_methods (id, code, name, is_active) VALUES ($1, $2, $3, true)`,
		globalMethodID, globalCode, globalName,
	)
	return err
}

// withCredentials reconstruye la connection string del contenedor reemplazando usuario y
// contraseña — evita depender de las APIs internas de testcontainers para host/puerto.
func withCredentials(connStr, user, password string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, password)
	return u.String(), nil
}

// countRaw cuenta filas bajo un RLSContext dado — usado para las aserciones de invisibilidad
// cross-tenant (nunca vía el WHERE tenant_id=... del repository, para no confundir aislamiento
// por RLS con un filtro de la query).
func countRaw(t *testing.T, rc postgres.RLSContext, query string, args ...interface{}) int {
	t.Helper()
	var count int
	err := postgres.WithRLSInTransaction(context.Background(), appDB, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&count)
	})
	require.NoError(t, err)
	return count
}

// insertEnablement inserta una fila de habilitación (tenant × global) bajo el RLSContext del
// tenant dado y devuelve su id. Corre por el mismo camino que la escritura real (una tx RLS con
// SET LOCAL app.tenant_id), así el WITH CHECK de la policy se ejerce de verdad.
func insertEnablement(t *testing.T, tenantID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := postgres.WithRLSInTransaction(context.Background(), appDB, postgres.RLSContext{TenantID: tenantID.String()},
		func(ctx context.Context, tx *sql.Tx) error {
			_, e := tx.ExecContext(ctx, `
				INSERT INTO payment_methods (id, tenant_id, global_payment_method_id, is_active)
				VALUES ($1, $2, $3, true)
			`, id, tenantID, globalMethodID)
			return e
		})
	require.NoError(t, err, "el INSERT de la propia fila del tenant debe pasar el WITH CHECK")
	return id
}

// containsID indica si la lista de métodos incluye ese id.
func containsID(list []*entity.PaymentMethod, id uuid.UUID) bool {
	for _, pm := range list {
		if pm.ID == id {
			return true
		}
	}
	return false
}

// TestPaymentMethods_CrossTenantIsolation_FailsClosed es el corazón de T6: prueba la cadena
// fail-closed sobre el plano tenant `payment_methods` bajo el rol NOBYPASSRLS.
func TestPaymentMethods_CrossTenantIsolation_FailsClosed(t *testing.T) {
	repo := repository.NewPostgresPaymentMethodRepository(appDB)

	tenantA := uuid.New()
	tenantB := uuid.New()

	// (a) Sembrar la fila de habilitación de A dentro de una tx con RLSContext{A}.
	rowID := insertEnablement(t, tenantA)

	t.Run("(b) A ve su fila por consulta cruda", func(t *testing.T) {
		count := countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM payment_methods WHERE id = $1`, rowID)
		require.Equal(t, 1, count)
	})

	t.Run("(b) A ve su fila por el repository, con code/name resueltos del catálogo global (JOIN bajo RLS)", func(t *testing.T) {
		got, err := repo.FindByID(context.Background(), rowID, tenantA)
		require.NoError(t, err)
		require.NotNil(t, got, "el propio tenant debe ver su fila")
		require.Equal(t, rowID, got.ID)
		require.Equal(t, globalCode, got.Code, "el JOIN al catálogo global (control-plane, sin RLS) debe resolver el code bajo RLS activo")
		require.Equal(t, globalName, got.Name)

		list, err := repo.FindAll(context.Background(), tenantA, false)
		require.NoError(t, err)
		require.True(t, containsID(list, rowID), "FindAll(A) debe incluir la fila de A")
	})

	t.Run("(c) B NO ve la fila de A por consulta cruda", func(t *testing.T) {
		count := countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM payment_methods WHERE id = $1`, rowID)
		require.Equal(t, 0, count, "la fila de tenantA es visible desde tenantB")
	})

	t.Run("(c) B NO ve la fila de A por el repository", func(t *testing.T) {
		got, err := repo.FindByID(context.Background(), rowID, tenantB)
		require.NoError(t, err)
		require.Nil(t, got, "tenantB no puede ver la fila de tenantA vía FindByID")

		list, err := repo.FindAll(context.Background(), tenantB, false)
		require.NoError(t, err)
		require.False(t, containsID(list, rowID), "FindAll(B) no puede incluir la fila de A")
	})

	t.Run("(d) INSERT con tenant_id=A bajo sesión de tenant B viola WITH CHECK", func(t *testing.T) {
		err := postgres.WithRLSInTransaction(context.Background(), appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, `
					INSERT INTO payment_methods (id, tenant_id, global_payment_method_id, is_active)
					VALUES ($1, $2, $3, true)
				`, uuid.New(), tenantA, globalMethodID)
				return err
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(e) UPDATE cross-tenant es no-op: B no puede togglear la fila de A", func(t *testing.T) {
		var affected int64
		err := postgres.WithRLSInTransaction(context.Background(), appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				res, e := tx.ExecContext(ctx,
					`UPDATE payment_methods SET is_active = false WHERE id = $1`, rowID)
				if e != nil {
					return e
				}
				affected, e = res.RowsAffected()
				return e
			})
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		// La fila de A sigue activa: el intento de B no tuvo efecto.
		got, err := repo.FindByID(context.Background(), rowID, tenantA)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.True(t, got.IsActive, "el is_active de A no debe haber cambiado por el UPDATE de B")
	})
}

// TestPaymentMethods_FailClosed_WithoutTenantContext prueba que el repository quedó 100% migrado
// a WithRLSInTransaction: una consulta corrida SIN fijar app.tenant_id (ni SET LOCAL ni tx RLS)
// debe devolver cero filas o errorar, nunca todas. Corre después del test de arriba a propósito:
// reutiliza el pool `appDB` cuyas conexiones físicas YA tuvieron SET LOCAL app.tenant_id fijado
// (y reseteado al commitear) — el escenario real de un pool reutilizado entre requests de
// distintos tenants, más representativo que una conexión nunca tocada.
func TestPaymentMethods_FailClosed_WithoutTenantContext(t *testing.T) {
	// Garantizar que hay al menos una fila para que "0 visibles" sea significativo.
	_ = insertEnablement(t, uuid.New())

	var count int
	err := appDB.QueryRowContext(context.Background(), `SELECT count(*) FROM payment_methods`).Scan(&count)
	if err != nil {
		return // fail-closed vía error (current_setting sin valor fijado): comportamiento esperado
	}
	require.Equal(t, 0, count, "payment_methods devolvió filas sin contexto de tenant fijado — RLS no está aislando")
}
